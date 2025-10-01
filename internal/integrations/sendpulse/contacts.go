package sendpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ContactListBody struct {
	Limit                int                    `json:"limit,omitempty"`
	Offset               int                    `json:"offset,omitempty"`
	From                 string                 `json:"from,omitempty"`
	To                   string                 `json:"to,omitempty"`
	UpdateFrom           string                 `json:"updateFrom,omitempty"`
	UpdateTo             string                 `json:"updateTo,omitempty"`
	FirstName            string                 `json:"firstName,omitempty"`
	LastName             string                 `json:"lastName,omitempty"`
	ResponsibleIds       []int                  `json:"responsibleIds,omitempty"`
	SourceType           []int                  `json:"sourceType,omitempty"`
	Phone                string                 `json:"phone,omitempty"`
	Email                string                 `json:"email,omitempty"`
	TagIds               []int                  `json:"tagIds,omitempty"`
	MessengerTypeIds     []int                  `json:"messengerTypeIds,omitempty"`
	MessengerLogin       string                 `json:"messengerLogin,omitempty"`
	SortBy               *SortBy                `json:"sortBy,omitempty"`
	Attributes           []*SearchAttributes    `json:"attributes,omitempty"`
	Location             string                 `json:"location,omitempty"`
	FieldValueConditions []FieldValueConditions `json:"fieldValueConditions,omitempty"`
	Ids                  []int                  `json:"ids,omitempty"`
}
type SortBy struct {
	Direction string `json:"direction,omitempty"`
	FieldName string `json:"fieldName,omitempty"`
}
type SearchAttributes struct {
	ID         string `json:"id,omitempty"`
	Expression string `json:"expression,omitempty"`
	Value      string `json:"value,omitempty"`
}
type FieldValueConditions struct {
	Field      string `json:"field,omitempty"`
	Expression string `json:"expression,omitempty"`
	Value      string `json:"value,omitempty"`
}

type Attachments struct {
	ID         int    `json:"id,omitempty"`
	Link       string `json:"link,omitempty"`
	EntityID   int    `json:"entityId,omitempty"`
	EntityType string `json:"entityType,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
}
type Comments struct {
	ID          int         `json:"id,omitempty"`
	UserID      int         `json:"userId,omitempty"`
	Text        string      `json:"text,omitempty"`
	CreatedAt   time.Time   `json:"createdAt,omitempty"`
	UpdatedAt   time.Time   `json:"updatedAt,omitempty"`
	Attachments Attachments `json:"attachments,omitempty"`
	ChildCount  int         `json:"childCount,omitempty"`
	ChildUsers  []int       `json:"childUsers,omitempty"`
}
type Tags struct {
	ID              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	ColorText       string `json:"colorText,omitempty"`
	ColorBackground string `json:"colorBackground,omitempty"`
	ContactCount    int    `json:"contactCount,omitempty"`
	TaskCount       int    `json:"taskCount,omitempty"`
}
type Phones struct {
	ID     int    `json:"id,omitempty"`
	Phone  string `json:"phone,omitempty"`
	IsMain bool   `json:"isMain,omitempty"`
}
type Emails struct {
	ID     int    `json:"id,omitempty"`
	Email  string `json:"email,omitempty"`
	IsMain bool   `json:"isMain,omitempty"`
}
type Messengers struct {
	ID            int    `json:"id,omitempty"`
	TypeID        int    `json:"typeId,omitempty"`
	Login         string `json:"login,omitempty"`
	BotID         string `json:"botId,omitempty"`
	ContactID     string `json:"contactId,omitempty"`
	Status        int    `json:"status,omitempty"`
	ChatbotURL    string `json:"chatbotUrl,omitempty"`
	IsMainChatbot bool   `json:"isMainChatbot,omitempty"`
}
type Value struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}
type Attributes struct {
	ID              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Status          int    `json:"status,omitempty"`
	Type            int    `json:"type,omitempty"`
	Mandatory       bool   `json:"mandatory,omitempty"`
	ContactCardShow bool   `json:"contactCardShow,omitempty"`
	Order           int    `json:"order,omitempty"`
	Options         []any  `json:"options,omitempty"`
	Value           Value  `json:"value,omitempty"`
	Default         bool   `json:"default,omitempty"`
}
type EventData struct {
	DealName string `json:"dealName,omitempty"`
}
type History struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"userId,omitempty"`
	ContactID int       `json:"contactId,omitempty"`
	EventType string    `json:"eventType,omitempty"`
	EventTime time.Time `json:"eventTime,omitempty"`
	EventData EventData `json:"eventData,omitempty"`
}
type List struct {
	ID                int           `json:"id,omitempty"`
	UserID            int           `json:"userId,omitempty"`
	SourceType        string        `json:"sourceType,omitempty"`
	ResponsibleID     int           `json:"responsibleId,omitempty"`
	FirstName         string        `json:"firstName,omitempty"`
	LastName          string        `json:"lastName,omitempty"`
	DealsQty          int           `json:"dealsQty,omitempty"`
	ExternalContactID string        `json:"externalContactId,omitempty"`
	Comments          []*Comments   `json:"comments,omitempty"`
	Tags              []*Tags       `json:"tags,omitempty"`
	Phones            []*Phones     `json:"phones,omitempty"`
	Emails            []*Emails     `json:"emails,omitempty"`
	Messengers        []*Messengers `json:"messengers,omitempty"`
	Attributes        []*Attributes `json:"attributes,omitempty"`
	History           []*History    `json:"history,omitempty"`
	Tasks             []int         `json:"tasks,omitempty"`
	CreatedAt         time.Time     `json:"createdAt,omitempty"`
	UpdatedAt         time.Time     `json:"updatedAt,omitempty"`
	Attachments       Attachments   `json:"attachments,omitempty"`
}
type ListData struct {
	List          []*List `json:"list,omitempty"`
	Total         int     `json:"total,omitempty"`
	SearchRequest string  `json:"searchRequest,omitempty"`
}

func (this *APIClient) GetContactList(ctx context.Context, search *ContactListBody) (*Response[*ListData], error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &Response[*ListData]{}
	body, err := json.Marshal(search)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, "/contacts/get-list").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx,
		req,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type CreateContactRequest struct {
	ResponsibleID     int    `json:"responsibleId"`
	FirstName         string `json:"firstName,omitempty"`
	LastName          string `json:"lastName,omitempty"`
	ExternalContactID string `json:"externalContactId,omitempty"`
}

type CreateContactResponse struct {
	Data *Contact `json:"data,omitempty"`
}

type Contact struct {
	ID                int                 `json:"id,omitempty"`
	UserID            int                 `json:"userId,omitempty"`
	SourceType        string              `json:"sourceType,omitempty"`
	ResponsibleID     int                 `json:"responsibleId,omitempty"`
	FirstName         string              `json:"firstName,omitempty"`
	LastName          string              `json:"lastName,omitempty"`
	DealsQty          int                 `json:"dealsQty,omitempty"`
	ExternalContactID string              `json:"externalContactId,omitempty"`
	Comments          []*ContactComment   `json:"comments,omitempty"`
	Tags              []*ContactTag       `json:"tags,omitempty"`
	Phones            []*ContactPhone     `json:"phones,omitempty"`
	Emails            []*ContactEmail     `json:"emails,omitempty"`
	Messengers        []*ContactMessenger `json:"messengers,omitempty"`
	Attributes        []*ContactAttribute `json:"attributes,omitempty"`
	History           []*ContactHistory   `json:"history,omitempty"`
	Tasks             []int               `json:"tasks,omitempty"`
	CreatedAt         time.Time           `json:"createdAt,omitempty"`
	UpdatedAt         time.Time           `json:"updatedAt,omitempty"`
	Attachments       *EntityAttachment   `json:"attachments,omitempty"`
}

type ContactComment struct {
	ID          int               `json:"id,omitempty"`
	UserID      int               `json:"userId,omitempty"`
	Text        string            `json:"text,omitempty"`
	CreatedAt   time.Time         `json:"createdAt,omitempty"`
	UpdatedAt   time.Time         `json:"updatedAt,omitempty"`
	Attachments *EntityAttachment `json:"attachments,omitempty"`
	ChildCount  int               `json:"childCount,omitempty"`
	ChildUsers  []int             `json:"childUsers,omitempty"`
}

type ContactTag struct {
	ID              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	ColorText       string `json:"colorText,omitempty"`
	ColorBackground string `json:"colorBackground,omitempty"`
	ContactCount    int    `json:"contactCount,omitempty"`
	TaskCount       int    `json:"taskCount,omitempty"`
}

type ContactPhone struct {
	ID     int    `json:"id,omitempty"`
	Phone  string `json:"phone,omitempty"`
	IsMain bool   `json:"isMain,omitempty"`
}

type ContactEmail struct {
	ID     int    `json:"id,omitempty"`
	Email  string `json:"email,omitempty"`
	IsMain bool   `json:"isMain,omitempty"`
}

type ContactMessenger struct {
	ID            int    `json:"id,omitempty"`
	TypeID        int    `json:"typeId,omitempty"`
	Login         string `json:"login,omitempty"`
	BotID         string `json:"botId,omitempty"`
	ContactID     string `json:"contactId,omitempty"`
	Status        int    `json:"status,omitempty"`
	ChatbotURL    string `json:"chatbotUrl,omitempty"`
	IsMainChatbot bool   `json:"isMainChatbot,omitempty"`
}

type ContactAttribute struct {
	ID              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Status          int    `json:"status,omitempty"`
	Type            int    `json:"type,omitempty"`
	Mandatory       bool   `json:"mandatory,omitempty"`
	ContactCardShow bool   `json:"contactCardShow,omitempty"`
	Order           int    `json:"order,omitempty"`
	Options         []any  `json:"options,omitempty"`
	Value           Value  `json:"value,omitempty"`
	Default         bool   `json:"default,omitempty"`
}

type ContactHistory struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"userId,omitempty"`
	ContactID int       `json:"contactId,omitempty"`
	EventType string    `json:"eventType,omitempty"`
	EventTime time.Time `json:"eventTime,omitempty"`
	EventData EventData `json:"eventData,omitempty"`
}

type EntityAttachment struct {
	ID         int       `json:"id,omitempty"`
	Link       string    `json:"link,omitempty"`
	EntityID   int       `json:"entityId,omitempty"`
	EntityType string    `json:"entityType,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}

// CreateContact creates a new contact
// https://sendpulse.com/integrations/api/crm#/Contacts/post_contacts_create
func (this *APIClient) CreateContact(ctx context.Context, request *CreateContactRequest) (*CreateContactResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &CreateContactResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, "/contacts/create").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type UpdateContactRequest struct {
	ResponsibleID int    `json:"responsibleId"`
	FirstName     string `json:"firstName,omitempty"`
	LastName      string `json:"lastName,omitempty"`
}

type UpdateContactResponse struct {
	Data *Contact `json:"data,omitempty"`
}

// UpdateContact updates information about the contact
// https://sendpulse.com/integrations/api/crm#/Contacts/put_contacts__contactId_
func (this *APIClient) UpdateContact(ctx context.Context, id int64, request *UpdateContactRequest) (*UpdateContactResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &UpdateContactResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%d", id)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type GetContactResponse struct {
	Data *Contact `json:"data,omitempty"`
}

// GetContact gets information about a contact by ID
// https://sendpulse.com/integrations/api/crm#/Contacts/get_contacts__contactId_
func (this *APIClient) GetContact(ctx context.Context, id int64) (*GetContactResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &GetContactResponse{}

	req := this.NewRequest(http.MethodGet, fmt.Sprintf("/contacts/%d", id)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteContact removes contact by ID
// https://sendpulse.com/integrations/api/crm#/Contacts/delete_contacts__contactId_
func (this *APIClient) DeleteContact(ctx context.Context, id int64) error {
	token, err := this.GetToken()
	if err != nil {
		return err
	}

	result := new(any) // delete endpoints typically return empty response

	req := this.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d", id)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
