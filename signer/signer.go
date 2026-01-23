package signer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/x893675/valhalla-common/utils/random"
)

func init() {
	_ = Register(defaultAlgorithm, sha256.New)
}

type SignatureAlgorithmFn func() hash.Hash

type signatureAlgorithms struct {
	algorithmsMap map[string]SignatureAlgorithmFn
}

var (
	_signatureAlgorithm = defaultSignatureAlgorithms()
)

var (
	ErrExist = errors.New("algorithms already exist")
)

func defaultSignatureAlgorithms() signatureAlgorithms {
	return signatureAlgorithms{algorithmsMap: map[string]SignatureAlgorithmFn{}}
}

func Register(name string, fn SignatureAlgorithmFn) error {
	return _signatureAlgorithm.registerComponent(name, fn)
}

func Load(kv string) (SignatureAlgorithmFn, bool) {
	return _signatureAlgorithm.load(kv)
}

func (h *signatureAlgorithms) load(kv string) (SignatureAlgorithmFn, bool) {
	c, exist := h.algorithmsMap[kv]
	return c, exist
}

func (h *signatureAlgorithms) registerComponent(name string, fn SignatureAlgorithmFn) error {
	_, exist := h.algorithmsMap[name]
	if exist {
		return ErrExist
	}
	h.algorithmsMap[name] = fn
	return nil
}

const (
	defaultAlgorithm  = "HMAC-SHA256"
	iso8601DateFormat = "20060102T150405Z"
	yyyymmdd          = "20060102"
)

const (
	queryKeySignature      = "Signature"
	queryKeyAlgorithm      = "SignatureAlgorithm"
	queryKeyCredential     = "AccessKey"
	queryKeyTimestamp      = "Timestamp"
	queryKeySignatureNonce = "SignatureNonce"
)

type Credential struct {
	Timestamp          string    `json:"timestamp" query:"Timestamp" form:"Timestamp" validate:"required"`
	SignatureAlgorithm string    `json:"signatureAlgorithm" query:"SignatureAlgorithm" form:"Timestamp" validate:"required"`
	SignatureNonce     string    `json:"signatureNonce" query:"SignatureNonce" form:"Timestamp" validate:"required"`
	Signature          string    `json:"signature" query:"Signature" form:"Timestamp" validate:"required"`
	AccessKey          string    `json:"accessKey" query:"AccessKey" form:"Timestamp" validate:"required"`
	AccessSecret       string    `json:"accessSecret"`
	TimestampTime      time.Time `json:"time"`
	AlgorithmFn        SignatureAlgorithmFn
}

var lf = []byte{'\n'}

func writeURI(r *http.Request, requestData io.Writer) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	//path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	_, _ = requestData.Write([]byte(path))
}

func writeQuery(r *http.Request, requestData io.Writer) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		if k == queryKeySignature {
			continue
		}
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{'&'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeBody(fn SignatureAlgorithmFn, r *http.Request, requestData io.StringWriter) {
	var b []byte
	// If the payload is empty, use the empty string as the input to the SHA256 function
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
	}
	_, _ = requestData.WriteString(hex.EncodeToString(gHash(fn(), b)))
}

func gHash(h hash.Hash, data []byte) []byte {
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func gHmac(fn SignatureAlgorithmFn, key, data []byte) []byte {
	h := hmac.New(fn, key)
	return gHash(h, data)
}

func NewAccessKeyAuth(accessKey, accessSecret string, algorithm string) *Credential {
	a := &Credential{
		SignatureNonce: random.RandStringBytesMaskImprSrcUnsafe(16),
		AccessKey:      accessKey,
		AccessSecret:   accessSecret,
		TimestampTime:  time.Now().UTC(),
	}
	a.Timestamp = a.TimestampTime.Format(iso8601DateFormat)
	fn, ok := Load(algorithm)
	if !ok {
		a.SignatureAlgorithm = defaultAlgorithm
		a.AlgorithmFn, _ = Load(defaultAlgorithm)
	} else {
		a.SignatureAlgorithm = algorithm
		a.AlgorithmFn = fn
	}
	return a
}

func NewAccessKeyAuthRequest(req *http.Request) (*Credential, error) {
	var err error
	uValues := req.URL.Query()
	a := &Credential{
		Timestamp:          uValues.Get(queryKeyTimestamp),
		SignatureAlgorithm: uValues.Get(queryKeyAlgorithm),
		SignatureNonce:     uValues.Get(queryKeySignatureNonce),
		Signature:          uValues.Get(queryKeySignature),
		AccessKey:          uValues.Get(queryKeyCredential),
		AccessSecret:       "",
	}
	if a.AccessKey == "" {
		return nil, fmt.Errorf("accesskey not found")
	}
	if a.Signature == "" {
		return nil, fmt.Errorf("signature not found")
	}
	if a.SignatureNonce == "" {
		return nil, fmt.Errorf("signature nonce not found")
	}
	a.TimestampTime, err = time.Parse(iso8601DateFormat, a.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("can not parse timestamp")
	}
	if a.SignatureAlgorithm == "" {
		a.SignatureAlgorithm = defaultAlgorithm
	}
	fn, ok := Load(a.SignatureAlgorithm)
	if !ok {
		return nil, fmt.Errorf("unsupport signature algorithm")
	}
	a.AlgorithmFn = fn

	return a, nil
}

func (a *Credential) CheckSignature(req *http.Request) error {
	result := a.stringToSign(req)
	if a.Signature != result {
		return fmt.Errorf("ak/sk signature check failed. expected: %s, got: %s", a.Signature, result)
	}
	return nil
}

func (a *Credential) SignRequest(req *http.Request) error {
	values := req.URL.Query()
	values.Set(queryKeyTimestamp, a.TimestampTime.Format(iso8601DateFormat))
	values.Set(queryKeyAlgorithm, a.SignatureAlgorithm)
	values.Set(queryKeyCredential, a.AccessKey)
	values.Set(queryKeySignatureNonce, a.SignatureNonce)
	req.URL.RawQuery = values.Encode()

	values = req.URL.Query()
	values.Set(queryKeySignature, a.stringToSign(req))
	req.URL.RawQuery = values.Encode()
	return nil
}

func (a *Credential) stringToSign(req *http.Request) string {
	lastData := bytes.NewBufferString(a.SignatureAlgorithm)
	lastData.Write(lf)
	lastData.Write([]byte(a.TimestampTime.Format(iso8601DateFormat)))
	lastData.Write(lf)
	lastData.WriteString(hex.EncodeToString(a.signRequest(req)))
	data := gHmac(a.AlgorithmFn, a.signKey(), lastData.Bytes())
	return hex.EncodeToString(data)
}

func (a *Credential) signKey() []byte {
	data := gHmac(a.AlgorithmFn, []byte(a.AccessSecret), []byte(a.TimestampTime.Format(yyyymmdd)))
	return gHmac(a.AlgorithmFn, data, []byte("request"))
}

func (a *Credential) signRequest(r *http.Request) []byte {
	requestData := bytes.NewBufferString("")

	requestData.Write([]byte(r.Method))
	requestData.Write(lf)

	writeURI(r, requestData)
	requestData.Write(lf)

	writeQuery(r, requestData)
	requestData.Write(lf)

	writeBody(a.AlgorithmFn, r, requestData)

	return gHash(a.AlgorithmFn(), requestData.Bytes())
}
