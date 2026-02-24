# Plano: Sistema de Persistência de API Keys com Expiração de 30 Dias

## Visão Geral
Implementar um sistema completo de gerenciamento de API keys com persistência, expiração automática de 30 dias e sistema de notificações para alertar usuários sobre chaves expiradas ou próximas de expirar.

## Estrutura Atual Analisada
- **Backend (Go)**: Configurações armazenadas em `config.json` via `internal/config/store.go`
- **Frontend (React)**: WebUI em `webui/src/` gerenciando API keys através de `/admin/keys`
- **Estrutura atual**: `Keys []string` em `internal/config/config.go`

## Fases de Implementação

### Fase 1: Extensão do Modelo de Dados
**Objetivo**: Adicionar campos de metadados para rastrear expiração das API keys

**Arquivos a modificar**:
- `internal/config/config.go`
  - Estender `Config` para incluir metadados de API keys
  - Criar novo struct `APIKeyMetadata` com:
    - `Key string` - a chave API
    - `CreatedAt time.Time` - data de criação
    - `ExpiresAt time.Time` - data de expiração (30 dias após criação)
    - `ID string` - identificador único da chave
  - Alterar `Keys []string` para `APIKeyMetadata []APIKeyMetadata` (ou adicionar novo campo para compatibilidade)

**Camada de compatibilidade**:
- Manter campo `Keys []string` para compatibilidade retroativa
- Implementar método `Keys()` que retorna apenas chaves válidas (não expiradas)
- Implementar método `AllAPIKeyMetadata()` que retorna todos os metadados

### Fase 2: Lógica de Persistência e Validação
**Objetivo**: Implementar persistência de metadados e validação de expiração

**Arquivos a criar/modificar**:
- `internal/config/apikey_manager.go` (novo)
  - `AddAPIKey(key string) error` - adiciona nova chave com metadados de expiração
  - `RemoveAPIKey(key string) error` - remove chave específica
  - `IsAPIKeyValid(key string) bool` - verifica se chave está válida e não expirada
  - `GetAPIKeyMetadata(key string) (APIKeyMetadata, bool)` - retorna metadados de uma chave
  - `GetExpiringKeys(daysBefore int) []APIKeyMetadata` - retorna chaves expirando em X dias
  - `GetExpiredKeys() []APIKeyMetadata` - retorna chaves já expiradas
  - `CleanExpiredKeys() int` - remove chaves expiradas (opcional, configurável)

- `internal/config/store.go`
  - Extender `Store` com índice para busca O(1) por chave
  - Atualizar métodos `Keys()`, `HasAPIKey()` para considerar expiração
  - Adicionar método `UpdateAPIKeyMetadata(key string, metadata APIKeyMetadata) error`
  - Adicionar método `GetAllAPIKeysMetadata() []APIKeyMetadata`

### Fase 3: Serviço de Monitoramento
**Objetivo**: Implementar verificação diária de expiração e disparo de notificações

**Arquivos a criar**:
- `internal/monitor/monitor.go` (novo)
  - `Monitor` struct com:
    - `store *config.Store`
    - `notifier *Notifier`
    - `checkInterval time.Duration` (24 horas por padrão)
    - `warningDays int` (7 dias por padrão)
  - `Start(ctx context.Context)` - inicia goroutine de monitoramento
  - `CheckNow()` - executa verificação imediata
  - `Stop()` - para monitoramento
  - `checkExpirations()` - verifica chaves expiradas e próximas de expirar

- `internal/monitor/notifier.go` (novo)
  - `Notifier` struct com canais para diferentes tipos de notificações
  - `Notification` struct com:
    - `Type string` - "warning" ou "expired"
    - `APIKey string` - chave afetada (parcialmente oculta)
    - `Message string` - mensagem formatada
    - `ExpiresAt time.Time` - data de expiração
  - `Subscribe() <-chan Notification` - retorna canal de notificações
  - `notifyExpiring(keys []APIKeyMetadata)` - envia notificações de alerta
  - `notifyExpired(keys []APIKeyMetadata)` - envia notificações de expiração

### Fase 4: Integração com Backend
**Objetivo**: Integrar monitoramento com o servidor e endpoints existentes

**Arquivos a modificar**:
- `internal/admin/handler.go`
  - Adicionar campos `Monitor *monitor.Monitor` e `Notifier *monitor.Notifier` ao `Handler`
  - Modificar `addKey` para registrar metadados de expiração
  - Modificar `deleteKey` para remover metadados correspondentes

- `internal/admin/handler_accounts_crud.go` (ou arquivo apropriado)
  - Atualizar endpoints para retornar metadados de expiração
  - Adicionar endpoint `GET /admin/keys/expiring` - lista chaves próximas de expirar
  - Adicionar endpoint `GET /admin/keys/expired` - lista chaves expiradas

- `app/handler.go` ou `internal/server/router.go`
  - Inicializar `Monitor` e `Notifier` ao iniciar servidor
  - Registrar rota WebSocket ou SSE para notificações em tempo real

- `internal/auth/request.go`
  - Modificar `Determine()` para rejeitar chaves expiradas
  - Adicionar log quando chave expirada for usada

### Fase 5: Interface de Notificações (Frontend)
**Objetivo**: Implementar UI para exibir notificações de expiração

**Arquivos a criar/modificar**:
- `webui/src/components/NotificationBanner.jsx` (novo)
  - Componente para exibir banners de alerta
  - Suporta múltiplos tipos: warning, error, info
  - Botão de dismiss para cada notificação
  - Auto-dismiss após período configurável

- `webui/src/features/account/ApiKeyExpiryPanel.jsx` (novo)
  - Painel mostrando status de expiração das API keys
  - Lista chaves válidas, expirando em 7 dias e expiradas
  - Visualização de dias restantes com indicador visual (verde/amarelo/vermelho)
  - Ações para gerar nova chave

- `webui/src/features/account/useApiKeyExpiry.js` (novo)
  - Hook para gerenciar dados de expiração
  - Polling periódico para verificar chaves expirando
  - Conexão WebSocket/SSE para notificações em tempo real

- `webui/src/features/account/ApiKeysPanel.jsx` (modificar)
  - Adicionar indicador visual de dias restantes para cada chave
  - Mostrar badge de "Expirando em X dias" ou "Expirada"
  - Integrar com `ApiKeyExpiryPanel`

- `webui/src/components/AccountManagerContainer.jsx` (modificar)
  - Adicionar estado para notificações de expiração
  - Renderizar `NotificationBanner` quando houver alertas

### Fase 6: Sistema de Notificações em Tempo Real
**Objetivo**: Implementar WebSocket ou SSE para推送 notificações

**Arquivos a criar/modificar**:
- `internal/admin/handler_notifications.go` (novo)
  - Endpoint `GET /admin/notifications` - retorna histórico de notificações
  - Endpoint `GET /admin/notifications/stream` - SSE stream para notificações em tempo real
  - Middleware para autenticação admin

- `webui/src/hooks/useNotificationStream.js` (novo)
  - Hook para conectar ao SSE stream de notificações
  - Reconexão automática em caso de desconexão
  - Tratamento de diferentes tipos de notificações

### Fase 7: Internacionalização
**Objetivo**: Adicionar textos de notificação em múltiplos idiomas

**Arquivos a modificar**:
- `webui/src/locales/en.json`
  - Adicionar chaves para notificações de expiração
  - Ex: "apiKey.expiringSoon", "apiKey.expired", "apiKey.daysLeft"

- `webui/src/locales/zh.json`
  - Traduções correspondentes em chinês

### Fase 8: Testes Unitários
**Objetivo**: Implementar testes para validar persistência, expiração e notificações

**Arquivos a criar**:
- `internal/config/apikey_manager_test.go`
  - Teste de criação de API key com metadados
  - Teste de validação de expiração (chave válida, expirando, expirada)
  - Teste de persistência em arquivo JSON
  - Teste de compatibilidade retroativa (migração de keys[] para metadados)

- `internal/monitor/monitor_test.go`
  - Teste de verificação periódica
  - Teste de detecção de chaves expirando em 7 dias
  - Teste de detecção de chaves expiradas
  - Teste de envio de notificações

- `internal/auth/request_test.go` (modificar)
  - Teste de rejeição de chaves expiradas
  - Teste de log de tentativa de uso de chave expirada

- `webui/src/features/account/__tests__/ApiKeyExpiryPanel.test.jsx`
  - Teste de renderização de painel de expiração
  - Teste de exibição de indicadores visuais
  - Teste de ações do usuário

### Fase 9: Migração de Dados
**Objetivo**: Migrar API keys existentes para o novo formato com metadados

**Arquivos a criar**:
- `internal/config/migration.go` (novo)
  - `MigrateV1ToV2(cfg *Config) error` - converte keys[] para APIKeyMetadata[]
  - Para keys existentes, definir `CreatedAt` como data atual ou data de modificação do arquivo config
  - Definir `ExpiresAt` como 30 dias após a criação

- `internal/config/store.go` (modificar)
  - Adicionar detecção automática de versão de config
  - Executar migração ao carregar config antigo
  - Salvar config migrado

### Fase 10: Documentação
**Objetivo**: Documentar novo sistema de persistência

**Arquivos a criar/modificar**:
- `API.md` ou `README.MD`
  - Adicionar seção sobre expiração de API keys
  - Explicar como o sistema funciona
  - Documentar endpoints novos
  - Explicar como configurar período de alerta (7 dias)

## Cronograma Estimado
- Fase 1-2: 2-3 dias (Modelo de dados e persistência)
- Fase 3-4: 2 dias (Monitoramento e integração backend)
- Fase 5-6: 2-3 dias (Frontend e notificações em tempo real)
- Fase 7: 0.5 dia (Internacionalização)
- Fase 8-9: 2 dias (Testes e migração)
- Fase 10: 0.5 dia (Documentação)

**Total estimado: 10-12 dias**

## Critérios de Sucesso
- ✅ API keys persistem por 30 dias após criação
- ✅ Múltiplas API keys podem coexistir sem perda de dados
- ✅ Sistema valida expiração automaticamente
- ✅ Notificações de alerta (7 dias antes) e expiração são enviadas
- ✅ UI exibe notificações visíveis (banners, pop-ups, console)
- ✅ Testes unitários cobrem persistência, expiração e notificações
- ✅ Migração de dados existentes funciona sem perdas
- ✅ Chaves expiradas são rejeitadas em requisições de API

## Riscos e Mitigações
| Risco | Mitigação |
|-------|-----------|
| Incompatibilidade com configs existentes | Manter compatibilidade retroativa com campo `Keys []string` |
| Perda de dados durante migração | Backup automático do config.json antes de migrar |
| Sobrecarga de notificações | Implementar rate limiting e debounce |
| Testes não cobrem todos os cenários | Testes de integração além de unitários |
| Performance impactada por validação | Implementar cache de validação e índices O(1) |

## Decisões Arquiteturais
1. **Armazenamento**: JSON em disco (config.json) - alinhado com arquitetura atual
2. **Notificações**: SSE (Server-Sent Events) para simplicidade - WebSocket é overkill para push unidirecional
3. **Validação**: O(1) usando map[string]struct{} para índice - performance ótima
4. **Migração**: Automática ao carregar config antigo - transparência para usuário
5. **Expiração**: 30 dias fixo - pode ser configurável via RuntimeConfig no futuro

## Próximos Passos Após Aprovação
1. Implementar Fase 1 (Extensão do Modelo de Dados)
2. Implementar Fase 2 (Lógica de Persistência e Validação)
3. Continuar seguindo cronograma
4. Testar cada fase antes de prosseguir
5. Documentar decisões tomadas durante implementação
