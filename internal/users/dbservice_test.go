package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/ergomake/e2e/testutils"
)

func TestDBUsersService_Save(t *testing.T) {
	tt := []struct {
		name  string
		setup []User
		arg   User
		state []User
	}{
		{
			name: "update email",
			setup: []User{{
				Email:    "email1",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "email2",
				Username: "username1",
				Provider: "github",
			},
			state: []User{{
				Email:    "email2",
				Username: "username1",
				Provider: "github",
			}},
		},
		{
			name: "update username",
			setup: []User{{
				Email:    "email1",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "email1",
				Username: "username2",
				Provider: "github",
			},
			state: []User{{
				Email:    "email1",
				Username: "username2",
				Provider: "github",
			}},
		},
		{
			name: "creates a user",
			setup: []User{{
				Email:    "email1",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "email2",
				Username: "username2",
				Provider: "github",
			},
			state: []User{
				{
					Email:    "email1",
					Username: "username1",
					Provider: "github",
				},
				{
					Email:    "email2",
					Username: "username2",
					Provider: "github",
				},
			},
		},
		{
			name: "creates a user when same email but different provider",
			setup: []User{{
				Email:    "email1",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "email1",
				Username: "username2",
				Provider: "gitlab",
			},
			state: []User{
				{
					Email:    "email1",
					Username: "username1",
					Provider: "github",
				},
				{
					Email:    "email1",
					Username: "username2",
					Provider: "gitlab",
				},
			},
		},
		{
			name: "creates a user when same username but different provider",
			setup: []User{{
				Email:    "email1",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "email2",
				Username: "username1",
				Provider: "gitlab",
			},
			state: []User{
				{
					Email:    "email1",
					Username: "username1",
					Provider: "github",
				},
				{
					Email:    "email2",
					Username: "username1",
					Provider: "gitlab",
				},
			},
		},
		{
			name: "creates a user when same username and email but different provider",
			setup: []User{{
				Email:    "email1",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "email1",
				Username: "username1",
				Provider: "gitlab",
			},
			state: []User{
				{
					Email:    "email1",
					Username: "username1",
					Provider: "github",
				},
				{
					Email:    "email1",
					Username: "username1",
					Provider: "gitlab",
				},
			},
		},
		{
			name: "can have multiple users with empty email",
			setup: []User{{
				Email:    "",
				Username: "username1",
				Provider: "github",
			}},
			arg: User{
				Email:    "",
				Username: "username2",
				Provider: "github",
			},
			state: []User{
				{
					Email:    "",
					Username: "username1",
					Provider: "github",
				},
				{
					Email:    "",
					Username: "username2",
					Provider: "github",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db := testutils.CreateRandomDB(t)
			service := NewDBUsersService(db)

			dbSetup := make([]databaseUser, len(tc.setup))
			for i, u := range tc.setup {
				dbSetup[i] = databaseUser{User: u}
			}
			err := db.Table("users").Save(dbSetup).Error
			require.NoError(t, err)

			err = service.Save(context.Background(), tc.arg)
			require.NoError(t, err)

			var dbState []databaseUser
			err = db.Table("users").Find(&dbState).Error
			require.NoError(t, err)

			state := make([]User, len(dbState))
			for i, d := range dbState {
				state[i] = d.User
			}

			assert.Equal(t, tc.state, state)
		})
	}
}
