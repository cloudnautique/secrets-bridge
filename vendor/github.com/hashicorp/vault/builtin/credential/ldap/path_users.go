package ldap

import (
	"strings"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathUsers(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `users/(?P<name>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the LDAP user.",
			},

			"groups": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Comma-separated list of additional groups associated with the user.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathUserDelete,
			logical.ReadOperation:   b.pathUserRead,
			logical.UpdateOperation:  b.pathUserWrite,
		},

		HelpSynopsis:    pathUserHelpSyn,
		HelpDescription: pathUserHelpDesc,
	}
}

func (b *backend) User(s logical.Storage, n string) (*UserEntry, error) {
	entry, err := s.Get("user/" + n)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result UserEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *backend) pathUserDelete(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete("user/" + d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathUserRead(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	user, err := b.User(req.Storage, d.Get("name").(string))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"groups": strings.Join(user.Groups, ","),
		},
	}, nil
}

func (b *backend) pathUserWrite(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	groups := strings.Split(d.Get("groups").(string), ",")
	for i, g := range groups {
		groups[i] = strings.TrimSpace(g)
	}

	// Store it
	entry, err := logical.StorageEntryJSON("user/"+name, &UserEntry{
		Groups: groups,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(entry); err != nil {
		return nil, err
	}

	return nil, nil
}

type UserEntry struct {
	Groups []string
}

const pathUserHelpSyn = `
Manage additional groups for users allowed to authenticate.
`

const pathUserHelpDesc = `
This endpoint allows you to create, read, update, and delete configuration
for LDAP users that are allowed to authenticate, in particular associating
additional groups to them.

Deleting a user will not revoke their auth. To do this, do a revoke on "login/<username>" for
the usernames you want revoked.
`
