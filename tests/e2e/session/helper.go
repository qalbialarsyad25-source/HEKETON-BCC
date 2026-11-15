package invitation_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	eventdto "main/internal/modules/event/dto"
	"main/internal/modules/invitation/domain"
	"main/internal/modules/invitation/dto"
	"main/internal/shared/authz"
	apperrors "main/internal/shared/errors"
	"main/tests/e2e/factory"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	basePath        = "/api/v1"
	eventsPath      = basePath + "/events"
	invitationsPath = basePath + "/invitations"
	invitationPath  = basePath + "/invitation"
)

// actor is a signed-in user of a test.
type actor struct {
	ID    string
	Email string
	Token string
}

// listBody is the body of the paginated lists: {"data": [...], "meta": {...}}.
type listBody struct {
	Data []dto.InvitationResponse `json:"data"`
	Meta struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		TotalItems int `json:"total_items"`
		TotalPages int `json:"total_pages"`
	} `json:"meta"`
}

// newActor creates a user and signs them in.
func newActor(t *testing.T, ts *suite.Suite) actor {
	t.Helper()

	user := factory.CreateUser(t, ts.DB, ts.Adapters.Hasher, authz.RoleUser)
	token, _ := factory.Login(t, ts.Router, user.Email, factory.Password)

	return actor{ID: fmt.Sprint(user.ID), Email: user.Email, Token: token}
}

// createEvent creates an event through the API. The creator becomes its coordinator.
// It returns the event id.
func createEvent(t *testing.T, ts *suite.Suite, owner actor) string {
	t.Helper()

	req := eventdto.CreateEventRequest{
		Title:       "Go Meetup",
		Description: "An evening of Go talks",
		Date:        time.Now().Add(48 * time.Hour),
		Location:    "Malang",
		MaxQuota:    50,
	}

	w := helper.Do(t, ts.Router, http.MethodPost, eventsPath, req, owner.Token)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	return helper.ParseJSON(t, w, helper.SuccessResponse[eventdto.EventResponse]{}).(helper.SuccessResponse[eventdto.EventResponse]).Data.ID
}

// myEventIDs returns the ids of the events the user is a committee member of.
func myEventIDs(t *testing.T, ts *suite.Suite, who actor) []string {
	t.Helper()

	w := helper.Do(t, ts.Router, http.MethodGet, eventsPath+"/my", nil, who.Token)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var body eventdto.EventListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	ids := make([]string, 0, len(body.Data))
	for _, event := range body.Data {
		ids = append(ids, event.ID)
	}

	return ids
}

// postInvitation sends POST /invitations and returns the raw answer.
func postInvitation(t *testing.T, ts *suite.Suite, from actor, eventID string, to actor, role domain.Role) *httptest.ResponseRecorder {
	t.Helper()

	req := dto.CreateInvitationRequest{EventID: eventID, ReceiverID: to.ID, Role: string(role)}

	return helper.Do(t, ts.Router, http.MethodPost, invitationsPath, req, from.Token)
}

// invite sends an invitation that must succeed.
func invite(t *testing.T, ts *suite.Suite, from actor, eventID string, to actor, role domain.Role) dto.InvitationResponse {
	t.Helper()

	w := postInvitation(t, ts, from, eventID, to, role)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	return parseInvitation(t, w)
}

// act calls POST /invitations/:id/<action> (accept, decline or cancel) and returns the raw answer.
func act(t *testing.T, ts *suite.Suite, action, invitationID string, who actor) *httptest.ResponseRecorder {
	t.Helper()

	return helper.Do(t, ts.Router, http.MethodPost, invitationsPath+"/"+invitationID+"/"+action, nil, who.Token)
}

// mustAct runs an action that must succeed.
func mustAct(t *testing.T, ts *suite.Suite, action, invitationID string, who actor) dto.InvitationResponse {
	t.Helper()

	w := act(t, ts, action, invitationID, who)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	return parseInvitation(t, w)
}

func getInvitation(t *testing.T, ts *suite.Suite, invitationID string, who actor) *httptest.ResponseRecorder {
	t.Helper()

	return helper.Do(t, ts.Router, http.MethodGet, invitationPath+"/"+invitationID, nil, who.Token)
}

func parseInvitation(t *testing.T, w *httptest.ResponseRecorder) dto.InvitationResponse {
	t.Helper()

	return helper.ParseJSON(t, w, helper.SuccessResponse[dto.InvitationResponse]{}).(helper.SuccessResponse[dto.InvitationResponse]).Data
}

func parseList(t *testing.T, w *httptest.ResponseRecorder) listBody {
	t.Helper()
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var body listBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	return body
}

func ids(list []dto.InvitationResponse) []string {
	out := make([]string, 0, len(list))
	for _, invitation := range list {
		out = append(out, invitation.ID)
	}

	return out
}

func assertSender(t *testing.T, got dto.SenderResponse, want actor) {
	t.Helper()
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, strings.ToLower(want.Email), strings.ToLower(got.Email))
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, want helper.ErrorResponse) {
	t.Helper()
	assert.Equal(t, wantStatus, w.Code)
	assert.Equal(t, want, helper.ParseJSON(t, w, helper.ErrorResponse{}))
}

func validation(issues ...apperrors.Issue) helper.ErrorResponse {
	return helper.ErrorFrom(apperrors.Validation(issues))
}
