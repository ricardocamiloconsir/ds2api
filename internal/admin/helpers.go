package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ds2api/internal/config"
	"ds2api/internal/util"
)

// writeJSON and intFrom are package-internal aliases for the shared util versions.
var writeJSON = util.WriteJSON
var intFrom = util.IntFrom

func reverseAccounts(a []config.Account) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func intFromQuery(r *http.Request, key string, d int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nilIfZero(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

func toStringSlice(v any) ([]string, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		out = append(out, strings.TrimSpace(fmt.Sprintf("%v", item)))
	}
	return out, true
}

func toAccount(m map[string]any) config.Account {
	return config.Account{
		Email:    fieldString(m, "email"),
		Mobile:   fieldString(m, "mobile"),
		Password: fieldString(m, "password"),
		Token:    fieldString(m, "token"),
	}
}

func fieldString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func statusOr(v int, d int) int {
	if v == 0 {
		return d
	}
	return v
}

func accountMatchesIdentifier(acc config.Account, identifier string) bool {
	id := strings.TrimSpace(identifier)

	decodedId, err := url.QueryUnescape(id)
	if err == nil {
		id = decodedId
	}

	email := strings.TrimSpace(acc.Email)
	mobile := strings.TrimSpace(acc.Mobile)
	accId := acc.Identifier()

	fmt.Printf("[MATCH] Comparing: id='%s' (decoded='%s'), email='%s', mobile='%s', accId='%s'\n", strings.TrimSpace(identifier), id, email, mobile, accId)

	if id == "" {
		fmt.Printf("[MATCH] ID is empty\n")
		return false
	}
	if email == id {
		fmt.Printf("[MATCH] Matched by email\n")
		return true
	}
	if mobile == id {
		fmt.Printf("[MATCH] Matched by mobile\n")
		return true
	}
	if accId == id {
		fmt.Printf("[MATCH] Matched by Identifier()\n")
		return true
	}
	fmt.Printf("[MATCH] No match found\n")
	return false
}

func findAccountByIdentifier(store ConfigStore, identifier string) (config.Account, bool) {
	id := strings.TrimSpace(identifier)
	if id == "" {
		return config.Account{}, false
	}
	if acc, ok := store.FindAccount(id); ok {
		return acc, true
	}
	accounts := store.Snapshot().Accounts
	for _, acc := range accounts {
		if accountMatchesIdentifier(acc, id) {
			return acc, true
		}
	}
	return config.Account{}, false
}
