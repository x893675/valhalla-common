package user

import "encoding/json"

type Info interface {
	GetName() string
	GetID() uint64
	GetDomain() string
	GetEmail() string
	GetPhone() string
	GetGroups() []string
	SetExtra(key string, values []string)
	GetExtra(key string) []string
}

type DefaultInfo struct {
	Name   string              `json:"name"`
	ID     uint64              `json:"id"`
	Domain string              `json:"domain"`
	Email  string              `json:"email"`
	Phone  string              `json:"phone"`
	Groups []string            `json:"groups,omitempty"`
	Extra  map[string][]string `json:"extra,omitempty"`
}

func (i DefaultInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

func (i *DefaultInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i)
}

func (i *DefaultInfo) GetName() string {
	return i.Name
}

func (i *DefaultInfo) GetID() uint64 {
	return i.ID
}

func (i *DefaultInfo) GetDomain() string {
	return i.Domain
}

func (i *DefaultInfo) GetEmail() string {
	return i.Email
}

func (i *DefaultInfo) GetPhone() string {
	return i.Phone
}

func (i *DefaultInfo) GetGroups() []string {
	return i.Groups
}

func (i *DefaultInfo) SetExtra(key string, values []string) {
	if i.Extra == nil {
		i.Extra = make(map[string][]string)
	}
	i.Extra[key] = values
}

func (i *DefaultInfo) GetExtra(key string) []string {
	if i.Extra == nil {
		return nil
	}
	return i.Extra[key]
}
