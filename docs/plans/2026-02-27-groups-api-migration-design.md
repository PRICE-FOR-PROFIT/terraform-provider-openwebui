# Groups API Migration Design

**Date:** 2026-02-27
**Status:** Approved

## Overview

The latest version of Open WebUI changed how group user membership is managed. The `/api/v1/groups/id/{id}` endpoint no longer returns or accepts `user_ids`. Instead, three new endpoints handle user management:

- **POST /api/v1/groups/id/{id}/users** - Get users in a group
- **POST /api/v1/groups/id/{id}/users/add** - Add users to a group
- **POST /api/v1/groups/id/{id}/users/remove** - Remove users from a group

This design migrates the Terraform provider to use these new endpoints while maintaining complete backward compatibility for existing Terraform configurations.

## Goals

1. Support the new Open WebUI API for group user management
2. Maintain 100% backward compatibility - no HCL changes required
3. Use surgical diff approach - only add/remove users that changed
4. Prevent user disruption during Terraform apply operations
5. No changes to Terraform schema or state file format

## Architecture

### Client Layer Changes

**File:** `internal/provider/client/groups/client.go`

Add three new methods:

```go
func (c *Client) GetUsers(id string) ([]User, error)
func (c *Client) AddUsers(id string, userIDs []string) error
func (c *Client) RemoveUsers(id string, userIDs []string) error
```

Modify existing method:
- `Update()` - Remove `user_ids` from request payload (API ignores it)

**File:** `internal/provider/client/groups/types.go`

Add new type:

```go
type User struct {
    ID             string   `json:"id"`
    Name           string   `json:"name"`
    Email          string   `json:"email"`
    Role           string   `json:"role"`
    StatusEmoji    *string  `json:"status_emoji"`
    StatusMessage  *string  `json:"status_message"`
    StatusExpiresAt *int64  `json:"status_expires_at"`
    Bio            *string  `json:"bio"`
    Groups         []string `json:"groups"`
    IsActive       bool     `json:"is_active"`
}
```

Note: The `User` struct is only used internally to parse API responses. The resource layer extracts just the IDs.

### Resource Layer Changes

**File:** `internal/provider/groups_resource.go`

No schema changes. Orchestration changes in CRUD methods:

**Create Operation:**
1. Call `client.Create()` with name/description
2. If `user_ids` provided: call `client.AddUsers(groupID, userIDs)`
3. Call `client.Update()` with permissions
4. Store all attributes in state

**Read Operation:**
1. Call `client.Get(id)` for metadata
2. Call `client.GetUsers(id)` for user list
3. Extract user IDs, sort for consistency
4. Update state

**Update Operation:**
1. Call `client.GetUsers(id)` to get current membership
2. Diff current vs plan:
   - `toRemove = currentIDs - planIDs`
   - `toAdd = planIDs - currentIDs`
3. If `toRemove` not empty: call `client.RemoveUsers(id, toRemove)`
4. If `toAdd` not empty: call `client.AddUsers(id, toAdd)`
5. Call `client.Update(id, group)` with name/description/permissions only
6. Update state

**Delete Operation:**
- No changes

## API Request/Response Structures

### GetUsers - POST /api/v1/groups/id/{id}/users

Request: Empty body

Response:
```json
[{
  "id": "user-uuid",
  "name": "User Name",
  "email": "email@example.com",
  "role": "admin",
  "groups": [],
  "is_active": true
}]
```

Extract only the `id` field from each object.

### AddUsers - POST /api/v1/groups/id/{id}/users/add

Request:
```json
{
  "user_ids": ["id1", "id2"]
}
```

Response: Updated GroupResponse

### RemoveUsers - POST /api/v1/groups/id/{id}/users/remove

Request:
```json
{
  "user_ids": ["id1", "id2"]
}
```

Response: Updated GroupResponse

### Update - POST /api/v1/groups/id/{id}/update

Request (modified):
```json
{
  "name": "Group Name",
  "description": "Description",
  "permissions": {...}
}
```

Note: `user_ids` field removed from payload.

## Error Handling

### Partial Failure Scenarios

1. **Create fails after AddUsers**: Group exists but incomplete. Error returned to Terraform - user can retry.

2. **Update partial failures**:
   - RemoveUsers fails: Return error, no users added (safe)
   - AddUsers fails after RemoveUsers: Return error, drift will be detected on next plan
   - Update fails after user changes: Metadata not updated, drift detected on next plan

3. **User doesn't exist**: API error surfaced to Terraform with clear message

4. **Empty user lists**: Skip RemoveUsers/AddUsers calls (no-op)

### HTTP Error Handling

- **404 on Read**: Remove from state (deleted outside Terraform)
- **401/403**: Return authentication error
- **4xx/5xx**: Return descriptive error with status code and endpoint
- **Success code**: All operations expect 200 status

## Testing Strategy

### Unit Tests (Client Layer)
- Mock HTTP responses for GetUsers, AddUsers, RemoveUsers
- Test success and error conditions (404, 500, malformed JSON)
- Verify correct request payloads and headers
- Verify Update() excludes user_ids

### Acceptance Tests
- Create group with and without user_ids
- Read verifies user_ids populated correctly
- Update: add users, remove users, change both
- Update: modify permissions without changing users
- Edge cases: empty list, all removed, all new
- Import existing groups with users
- Drift detection when users changed outside Terraform

### Manual Testing
- Test against live Open WebUI instance
- Verify no user disruption during updates
- Verify member_count matches user list length

## Backward Compatibility

- Schema unchanged: `user_ids` remains `types.List` of strings
- HCL syntax unchanged: users still specify `user_ids = ["id1", "id2"]`
- State file format unchanged
- Existing Terraform configurations work without modification

## Implementation Notes

1. The surgical diff approach minimizes API calls and prevents unnecessary user disruption
2. User sorting ensures consistent state and prevents false drift detection
3. The GetUsers endpoint returns full user objects, but we only store IDs in state
4. The Update endpoint now focuses solely on group metadata (name, description, permissions)
5. Error messages should be descriptive to help users debug issues (e.g., "user ID xyz does not exist")

## Migration Path

Since this is a breaking API change and backward compatibility is not required:

1. Deploy updated provider alongside Open WebUI upgrade
2. Existing Terraform state will continue to work
3. Next `terraform plan` will show no changes (transparent migration)
4. No user action required

## Success Criteria

1. Provider works with new Open WebUI API
2. Existing Terraform configurations require no changes
3. User membership updates only affect changed users
4. All acceptance tests pass
5. No user disruption during terraform apply
