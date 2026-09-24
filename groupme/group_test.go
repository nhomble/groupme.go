package groupme

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestUpdateGroupCommandOmitsUnsetFields(t *testing.T) {
	name := "new-name"
	cmd := UpdateGroupCommand{
		Name: &name,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	marshaled := string(data)

	if !strings.Contains(marshaled, `"name":"new-name"`) {
		t.Errorf("expected marshaled command to include name, got %s", marshaled)
	}
	if strings.Contains(marshaled, `"share"`) {
		t.Errorf("expected marshaled command to omit share, got %s", marshaled)
	}
	if strings.Contains(marshaled, `"office_mode"`) {
		t.Errorf("expected marshaled command to omit office_mode, got %s", marshaled)
	}
	if strings.Contains(marshaled, `"image_url"`) {
		t.Errorf("expected marshaled command to omit image_url, got %s", marshaled)
	}
}

func TestJoinEscapesShareURLAndGroupID(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedPath string
	httpmock.RegisterResponder("POST", `=~^https://api\.groupme\.com/v3/groups/.*`,
		func(req *http.Request) (*http.Response, error) {
			capturedPath = req.URL.EscapedPath()
			return httpmock.NewStringResponse(200, `{"response":{}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Groups.Join("group/id", "share url/with space")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "/v3/groups/group%2Fid/join/share%20url%2Fwith%20space"
	if capturedPath != expected {
		t.Errorf("expected escaped path %q, got %q", expected, capturedPath)
	}

	if httpmock.GetTotalCallCount() != 1 {
		t.Errorf("Did not mock join group")
	}
}

func TestGetGroupEscapesId(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedPath string
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/.*`,
		func(req *http.Request) (*http.Response, error) {
			capturedPath = req.URL.EscapedPath()
			return httpmock.NewStringResponse(200, `{"response":{}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Groups.Get("weird/id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "/v3/groups/weird%2Fid"
	if capturedPath != expected {
		t.Errorf("expected escaped path %q, got %q", expected, capturedPath)
	}
}

func TestFindGroups(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups`,
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			return httpmock.NewStringResponse(200, `{"response":[{"id":"1","name":"Group One"},{"id":"2","name":"Group Two"}]}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	groups, err := client.Groups.Find(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "GET" {
		t.Errorf("expected GET, got %s", capturedMethod)
	}
	if capturedPath != "/v3/groups" {
		t.Errorf("expected path /v3/groups, got %s", capturedPath)
	}
	if len(groups) != 2 || groups[0].Name != "Group One" || groups[1].Name != "Group Two" {
		t.Errorf("unexpected groups result: %+v", groups)
	}
}

func TestFindGroupsAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups`,
		httpmock.NewStringResponder(500, `{"meta":{"errors":["boom"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	groups, err := client.Groups.Find(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if groups != nil {
		t.Errorf("expected nil groups on error, got %v", groups)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected StatusCode=500, got %d", apiErr.StatusCode)
	}
}

func TestCreateGroup(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	var capturedBody CreateGroupCommand
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			data, _ := io.ReadAll(req.Body)
			if err := json.Unmarshal(data, &capturedBody); err != nil {
				t.Fatalf("failed to unmarshal create body: %v", err)
			}
			return httpmock.NewStringResponse(200, `{"response":{"id":"new-id","name":"New Group"}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	group, err := client.Groups.Create(CreateGroupCommand{Name: "New Group", Share: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected POST, got %s", capturedMethod)
	}
	if capturedPath != "/v3/groups" {
		t.Errorf("expected path /v3/groups, got %s", capturedPath)
	}
	if capturedBody.Name != "New Group" || !capturedBody.Share {
		t.Errorf("unexpected request body: %+v", capturedBody)
	}
	if group == nil || group.ID != "new-id" || group.Name != "New Group" {
		t.Errorf("unexpected created group: %+v", group)
	}
}

func TestCreateGroupAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups",
		httpmock.NewStringResponder(400, `{"meta":{"errors":["invalid name"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	group, err := client.Groups.Create(CreateGroupCommand{Name: ""})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if group != nil {
		t.Errorf("expected nil group on error, got %v", group)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected StatusCode=400, got %d", apiErr.StatusCode)
	}
}

func TestUpdateGroup(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	var capturedBody UpdateGroupCommand
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups/group-1/update",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			data, _ := io.ReadAll(req.Body)
			if err := json.Unmarshal(data, &capturedBody); err != nil {
				t.Fatalf("failed to unmarshal update body: %v", err)
			}
			return httpmock.NewStringResponse(200, `{"response":{"id":"group-1","name":"Renamed"}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	name := "Renamed"
	group, err := client.Groups.Update("group-1", UpdateGroupCommand{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected POST, got %s", capturedMethod)
	}
	if capturedPath != "/v3/groups/group-1/update" {
		t.Errorf("expected path /v3/groups/group-1/update, got %s", capturedPath)
	}
	if capturedBody.Name == nil || *capturedBody.Name != "Renamed" {
		t.Errorf("unexpected request body: %+v", capturedBody)
	}
	if group == nil || group.Name != "Renamed" {
		t.Errorf("unexpected updated group: %+v", group)
	}
}

func TestUpdateGroupAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups/group-1/update",
		httpmock.NewStringResponder(403, `{"meta":{"errors":["forbidden"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	name := "Renamed"
	group, err := client.Groups.Update("group-1", UpdateGroupCommand{Name: &name})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if group != nil {
		t.Errorf("expected nil group on error, got %v", group)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected StatusCode=403, got %d", apiErr.StatusCode)
	}
}

func TestDeleteGroup(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups/group-1/destroy",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			return httpmock.NewStringResponse(200, `{}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	if err := client.Groups.Delete("group-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected POST, got %s", capturedMethod)
	}
	if capturedPath != "/v3/groups/group-1/destroy" {
		t.Errorf("expected path /v3/groups/group-1/destroy, got %s", capturedPath)
	}
}

func TestDeleteGroupAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups/group-1/destroy",
		httpmock.NewStringResponder(404, `{"meta":{"errors":["not found"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	err := client.Groups.Delete("group-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected StatusCode=404, got %d", apiErr.StatusCode)
	}
}

func TestReJoinGroup(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	var capturedBody struct {
		ID string `json:"group_id"`
	}
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups/join",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			data, _ := io.ReadAll(req.Body)
			if err := json.Unmarshal(data, &capturedBody); err != nil {
				t.Fatalf("failed to unmarshal rejoin body: %v", err)
			}
			return httpmock.NewStringResponse(200, `{"response":{"id":"group-1","name":"Rejoined Group"}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	group, err := client.Groups.ReJoin("group-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected POST, got %s", capturedMethod)
	}
	if capturedPath != "/v3/groups/join" {
		t.Errorf("expected path /v3/groups/join, got %s", capturedPath)
	}
	if capturedBody.ID != "group-1" {
		t.Errorf("expected group_id=group-1 in body, got %q", capturedBody.ID)
	}
	if group == nil || group.Name != "Rejoined Group" {
		t.Errorf("unexpected rejoined group: %+v", group)
	}
}

func TestReJoinGroupAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/groups/join",
		httpmock.NewStringResponder(400, `{"meta":{"errors":["cannot rejoin"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	group, err := client.Groups.ReJoin("group-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if group != nil {
		t.Errorf("expected nil group on error, got %v", group)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected StatusCode=400, got %d", apiErr.StatusCode)
	}
}

func TestDefaultGroupQueryReturnsIndependentCopies(t *testing.T) {
	first := DefaultGroupQuery()
	first.Page = 999
	first.Omit = append(first.Omit, "mutated")

	second := DefaultGroupQuery()
	if second.Page == 999 {
		t.Errorf("expected second call's Page to be unaffected by mutation of first, got %d", second.Page)
	}
	for _, v := range second.Omit {
		if v == "mutated" {
			t.Errorf("expected second call's Omit to be unaffected by mutation of first, got %v", second.Omit)
		}
	}
}
