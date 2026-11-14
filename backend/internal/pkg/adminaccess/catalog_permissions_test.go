package adminaccess

import "testing"

func TestResetHistoryRequiresSubscriptionReadPermission(t *testing.T) {
	path := "/api/v1/admin/subscriptions/:id/reset-history"
	if !RouteAllowed(Policy{GlobalPermissions: []string{"subscriptions.read"}}, "GET", path) {
		t.Fatal("subscription readers must be able to read reset history")
	}
	if RouteAllowed(Policy{GlobalPermissions: []string{"subscriptions.reset"}}, "GET", path) {
		t.Fatal("write-only administrators must not read reset history")
	}
}

func TestScopedPermissionCatalogRequiresExplicitGrant(t *testing.T) {
	for _, id := range []string{"accounts.identity.read", "accounts.usage.read", "accounts.options.write", "accounts.rate.write", "groups.list.read", "groups.capacity.read", "openai.tokens.refresh"} {
		found := false
		for _, entry := range PermissionCatalog() {
			if entry.ID == id {
				found = true
			}
		}
		if !found {
			t.Errorf("权限未登记: %s", id)
		}
	}
	if RouteAllowed(Policy{}, "POST", "/api/v1/admin/openai/refresh-token") {
		t.Fatal("旧策略不得继承新增权限")
	}
}
