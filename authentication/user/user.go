package user

import "encoding/json"

type UserType string

func (u UserType) String() string {
	return string(u)
}

const (
	UserTypeAccount        UserType = "account"
	UserTypeUser           UserType = "user"
	UserTypeAdmin          UserType = "admin"
	UserTypeService        UserType = "service"
	UserTypeServiceAccount UserType = "service_account"
)

type Info interface {
	UserType() UserType
	GetName() string
	GetID() string
	GetDomain() string
	GetEmail() string
	GetPhone() string
	GetGroups() []string
	SetExtra(key string, value any)
	GetExtra(key string) any
}

var _ Info = (*DefaultInfo)(nil)

type DefaultInfo struct {
	Type   UserType       `json:"type"`
	Name   string         `json:"name"`
	ID     string         `json:"id"`
	Domain string         `json:"domain"`
	Email  string         `json:"email"`
	Phone  string         `json:"phone"`
	Groups []string       `json:"groups,omitempty"`
	Extra  map[string]any `json:"extra,omitempty"`
}

func (i DefaultInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

func (i *DefaultInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i)
}

func (i *DefaultInfo) UserType() UserType {
	return i.Type
}

func (i *DefaultInfo) GetName() string {
	return i.Name
}

func (i *DefaultInfo) GetID() string {
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

func (i *DefaultInfo) SetExtra(key string, value any) {
	if i.Extra == nil {
		i.Extra = make(map[string]any)
	}
	i.Extra[key] = value
}

func (i *DefaultInfo) GetExtra(key string) any {
	if i.Extra == nil {
		return nil
	}
	return i.Extra[key]
}
