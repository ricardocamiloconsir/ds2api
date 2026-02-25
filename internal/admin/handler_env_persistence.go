package admin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ds2api/internal/config"
)

// ensureEnvBackedConfigPersistence persists runtime config updates when the store
// is backed by environment variables (e.g. Vercel).
func (h *Handler) ensureEnvBackedConfigPersistence(ctx context.Context) error {
	if !h.Store.IsEnvBacked() || !config.IsVercel() {
		return nil
	}

	vercelToken := strings.TrimSpace(os.Getenv("VERCEL_TOKEN"))
	projectID := strings.TrimSpace(os.Getenv("VERCEL_PROJECT_ID"))
	teamID := strings.TrimSpace(os.Getenv("VERCEL_TEAM_ID"))
	if vercelToken == "" || projectID == "" {
		return fmt.Errorf("configuração em modo env detectada, mas VERCEL_TOKEN/VERCEL_PROJECT_ID não estão configurados; use Vercel Sync para persistir")
	}

	_, cfgB64, err := h.Store.ExportJSONAndBase64()
	if err != nil {
		return fmt.Errorf("falha ao exportar configuração atual: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	params := buildVercelParams(teamID)
	headers := map[string]string{"Authorization": "Bearer " + vercelToken}
	envResp, status, err := vercelRequest(ctx, client, http.MethodGet, "https://api.vercel.com/v9/projects/"+projectID+"/env", params, headers, nil)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("falha ao consultar ambiente Vercel (status=%d): %w", status, err)
	}
	envs, _ := envResp["envs"].([]any)
	status, err = upsertVercelEnv(ctx, client, projectID, params, headers, envs, "DS2API_CONFIG_JSON", cfgB64)
	if err != nil || (status != http.StatusOK && status != http.StatusCreated) {
		return fmt.Errorf("falha ao salvar DS2API_CONFIG_JSON no Vercel (status=%d): %w", status, err)
	}

	manual, _ := triggerVercelDeployment(ctx, client, projectID, params, headers)
	if manual {
		return fmt.Errorf("configuração salva no Vercel, mas o redeploy automático falhou; faça deploy manual para aplicar")
	}

	return h.Store.SetVercelSync(h.computeSyncHash(), time.Now().Unix())
}
