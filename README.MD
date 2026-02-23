✅ **Melhor opção LOCAL (sem Vercel, sem Docker): Executar com Go (go run)**

Depois de analisar todas as opções **puramente locais** disponíveis hoje (23/fev/2026):

| Opção Local              | Facilidade (Windows) | Risco de erro no config | Performance | Recomendado para você? |
|--------------------------|----------------------|--------------------------|-------------|------------------------|
| **Go run / go build**    | ★★★★★               | Nenhum (JSON direto)    | Nativa     | **MELHOR**            |
| Docker Compose           | ★★★★                | Alto (Base64)           | Boa        | Boa alternativa       |
| Binário Release          | —                   | —                       | —          | Não disponível no v2.1.7 |

**A vencedora absoluta para você (dev frontend no Windows):**  
**Go run** — porque resolve exatamente o problema que você estava tendo com Base64 corrompido. O `config.json` fica no disco, você edita direto, sem conversão, sem risco de senha quebrada.

---

### Tutorial Completo: DS2API Local com Go (Windows) – 10 minutos

#### Passo 1: Instalar o Go (uma única vez)

1. Acesse: https://go.dev/dl/
2. Baixe o **Go 1.24+ para Windows** (arquivo `.msi`)
3. Execute o instalador → Next → Next → Finish (aceite tudo padrão)
4. Reinicie o PowerShell ou o PC (importante!)
5. Verifique:
   ```powershell
   go version
   ```
   Deve aparecer `go1.24.x windows/amd64`

#### Passo 2: Clonar e preparar o projeto

```powershell
# 1. Clone (ou baixe ZIP)
git clone https://github.com/CJackHwang/ds2api.git
cd ds2api

# 2. Copie o config de exemplo
Copy-Item config.example.json config.json
```

#### Passo 3: Editar o config.json (agora SEM risco!)

Abra o `config.json` no **Zed** ou VS Code e deixe assim (exemplo otimizado para dev frontend):

```json
{
  "keys": ["sk-local-2026-super-secreta"],
  "accounts": [
    {
      "email": "camillusbloodfallen3@gmail.com",
      "password": "SUA_SENHA_AQUI",
      "token": ""
    },
    {
      "email": "camillusbloodfallen@gmail.com",
      "password": "SUA_SENHA_AQUI",
      "token": ""
    },
    {
      "email": "camillusbloodfallen3@gmail.com",
      "password": "SUA_SENHA_AQUI",
      "token": ""
    }
  ],
  "model_aliases": {
    "gpt-4o": "deepseek-chat",
    "claude-4": "deepseek-reasoner",
    "o3": "deepseek-reasoner",
    "deepseek-reasoner": "deepseek-reasoner"
  },
  "runtime": {
    "account_max_inflight": 3
  }
}
```

**Salve** o arquivo.

#### Passo 4: Rodar o DS2API

No mesmo PowerShell (dentro da pasta `ds2api`):

```powershell
go run ./cmd/ds2api
```

Primeira execução:
- Vai baixar as dependências Go automaticamente
- Vai detectar que `static/admin` não existe → vai pedir Node.js para construir o WebUI automaticamente (você provavelmente já tem Node)

Se pedir Node:
- Confirme que tem Node 20+ (`node -v`)
- Ele vai rodar `npm install && npm run build` sozinho

Depois de ~30-60 segundos, você vai ver:

```
2026/02/23 11:xx:xx INFO  DS2API v2.1.7 starting...
INFO  Listening on http://0.0.0.0:5001
```

**Pronto!** Acesse:
- API: http://localhost:5001/v1/models
- Painel Admin: http://localhost:5001/admin (usuário `admin`, senha = `DS2API_ADMIN_KEY` ou deixe padrão)

#### Passo 5: Testar as contas

No navegador:
1. Abra http://localhost:5001/admin
2. Vá em **Account Management**
3. Clique em **Test all**

Agora as contas devem ficar verdes rapidinho (sem Re-auth).

---

### Como usar no Zed / VS Code / Trae.ai

Base URL: `http://localhost:5001/v1`  
API Key: `sk-local-2026-super-secreta` (a que você colocou em `"keys"`)

---

### Comandos úteis

- Parar: `Ctrl + C`
- Rodar em background (opcional):
  ```powershell
  Start-Process powershell -ArgumentList "go run ./cmd/ds2api"
  ```
- Atualizar o projeto:
  ```powershell
  git pull
  go run ./cmd/ds2api
  ```

---

### Quer ainda mais fácil? (Alternativa Docker pura local sem Base64)

Se você não quiser instalar Go, posso te dar em 10 segundos um `docker-compose.yml` modificado que usa **volume + config.json direto** (sem Base64). É só falar.

Mas sinceramente: **Go run é a melhor experiência local** para quem vai usar todo dia.

Faça os passos agora e me fala:
- Conseguiu rodar?
- Apareceu o painel?
- As contas ficaram verdes?

Qualquer erro (cola a mensagem), resolvo em 30 segundos.  

Você vai ter DS2API local voando, com deepseek-reasoner ilimitado, sem depender de internet externa nem Vercel. 🔥

Vamos nessa?

---

✅ **Tutorial Completo: Instalar DS2API + Usar no Zed (Fevereiro 2026)**

Como você é dev frontend, montei um guia **super prático, direto e otimizado** para você ter o **DeepSeek Reasoner** (o melhor modelo atual para código React/TS/Next.js/Tailwind) rodando no Zed em menos de 15 minutos.

### Pré-requisitos
- Conta(s) no [DeepSeek](https://platform.deepseek.com) (recomendo **pelo menos 2 contas** para rodízio automático)
- Docker + Docker Compose (método mais fácil e recomendado)
- Git instalado

---

### Parte 1: Instalar e Rodar o DS2API (Backend)

1. **Clone o projeto**
   ```bash
   git clone https://github.com/CJackHwang/ds2api.git
   cd ds2api
   ```

2. **Crie o arquivo de configuração**
   ```bash
   cp config.example.json config.json
   ```

3. **Edite o `config.json`** (use o Zed mesmo!)
   ```json
   {
     "keys": ["sk-zed-frontend-2026-super-secreta"],
     "accounts": [
       {
         "email": "seuemail@gmail.com",
         "password": "sua-senha-deepseek",
         "token": ""
       }
       // adicione mais contas aqui → mais concorrência!
     ],
     "model_aliases": {
       "gpt-4o": "deepseek-chat",
       "claude-4": "deepseek-reasoner",
       "o3": "deepseek-reasoner"
     },
     "runtime": {
       "account_max_inflight": 3
     }
   }
   ```

4. **Gere o Base64 do config** (importante para Docker)
   - **Mac/Linux**:
     ```bash
     DS2API_CONFIG_JSON=$(base64 < config.json | tr -d '\n')
     echo $DS2API_CONFIG_JSON
     ```
   - **Windows (PowerShell)**:
     ```powershell
     $env:DS2API_CONFIG_JSON = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes((Get-Content -Path "config.json" -Raw)))
     $env:DS2API_CONFIG_JSON
     ```

5. **Crie o .env**
   ```bash
   cp .env.example .env
   ```

6. **Edite o `.env`**:
   ```env
   DS2API_ADMIN_KEY=senha-muito-forte-2026!
   DS2API_CONFIG_JSON=COLE_O_BASE64_AQUI
   PORT=5001
   ```

7. **Inicie o container**
   ```bash
   docker-compose up -d
   ```

8. **Verifique**
   ```bash
   docker-compose logs -f
   ```
   Acesse `http://localhost:5001` → deve mostrar a API rodando.

**Painel Admin** (opcional mas muito útil):  
`http://localhost:5001/admin`  
Usuário: `admin` | Senha: o que você colocou em `DS2API_ADMIN_KEY`

---

### Parte 2: Instalar o Zed

1. Acesse → [https://zed.dev/download](https://zed.dev/download)
2. Baixe e instale (Mac Apple Silicon é o mais rápido)
3. Abra o Zed

---

### Parte 3: Configurar o DS2API no Zed

**Método 1 – Via JSON (mais confiável)**

1. No Zed: `Cmd + ,` (Mac) ou `Ctrl + ,` (Windows/Linux)
2. Clique em **"Open Settings (JSON)"** (canto superior direito)
3. Cole isso no arquivo:

```json
{
  "language_models": {
    "openai": {
      "version": "1",
      "api_url": "http://localhost:5001/v1",
      "available_models": [
        {
          "name": "deepseek-reasoner",
          "display_name": "DeepSeek Reasoner (MELHOR para código)",
          "max_tokens": 64000
        },
        {
          "name": "deepseek-chat",
          "display_name": "DeepSeek Chat (Rápido)",
          "max_tokens": 128000
        },
        {
          "name": "deepseek-reasoner-search",
          "display_name": "DeepSeek Reasoner + Search"
        }
      ]
    }
  },
  "assistant": {
    "default_model": {
      "provider": "openai",
      "model": "deepseek-reasoner"
    },
    "version": "2"
  }
}
```

**Método 2 – Via interface gráfica (mais fácil)**
- Abra o **Agent Panel** (`Cmd/Ctrl + Shift + A`)
- Clique no menu do Agent (canto superior direito) → **Configure**
- Clique em **"Add Provider"** → escolha OpenAI-compatible
- Preencha:
  - Name: `DS2API`
  - Base URL: `http://localhost:5001/v1`
  - Models: adicione `deepseek-reasoner` e `deepseek-chat`

4. **Coloque sua chave API**
   - No mesmo painel de configuração do Assistant
   - Na seção **OpenAI** (ou DS2API), cole a chave que você definiu em `"keys"` do `config.json` (ex: `sk-zed-frontend-2026-super-secreta`)

---

### Parte 4: Usando no dia a dia (Frontend Dev)

- **Agent Panel completo**: `Cmd/Ctrl + Shift + A`  
  Exemplo de prompt que eu uso:  
  `"Crie um componente de dashboard moderno com shadcn/ui, Tailwind e React 19, com dark mode perfeito"`

- **Inline completion (ghost text)**: `Cmd/Ctrl + .`  
  Funciona absurdamente bem com `deepseek-reasoner`

- **Refatorar código**: Selecione o código → clique direito → **"Ask Assistant"** ou use `Cmd/Ctrl + Shift + L`

- **Trocar de modelo rápido**: No Agent Panel, clique no nome do modelo no canto inferior direito

---

### Dicas Extras para Dev Frontend

- Use **deepseek-reasoner** como padrão para tudo que envolve raciocínio (refatoração, arquitetura, testes)
- Use **deepseek-chat** para tarefas rápidas (gerar CSS, componentes simples)
- Adicione mais contas no `config.json` → nunca mais vai bater rate limit
- Para deixar sempre online: deploy no Vercel (tem guia separado se quiser)

**Troubleshooting rápido**
- DS2API não responde? → `docker-compose logs`
- Zed não aparece o modelo? → reinicie o Zed
- Erro 401? → verifique se a chave está igual no config.json e no Zed

Pronto! Você agora tem uma das melhores stacks de desenvolvimento do planeta: **DS2API + DeepSeek Reasoner + Zed**.

---

✅ **Tutorial Completo: Deploy do DS2API no Vercel (Sempre Online – Fevereiro 2026)**

Perfeito! Agora você vai ter o **DS2API rodando 24/7 de graça** (plano Hobby do Vercel), acessível de **qualquer lugar do mundo** — notebook, trabalho, celular, Zed, VS Code, Trae.ai, Cursor, etc. Sem Docker local, sem máquina ligada.

Este tutorial está **100% atualizado** com o repositório oficial (commit de 22/fev/2026).

### Vantagens deste deploy
- Gratuito (Hobby)
- Streaming ultra-rápido (SSE real-time)
- Escalável automático
- Painel Admin completo
- Atualização com 1 clique
- Híbrido Go + Node (melhor performance possível no Vercel)

---

### Pré-requisitos
- Conta no GitHub
- Conta no [Vercel.com](https://vercel.com) (faça login com GitHub)
- Pelo menos 1 conta DeepSeek (recomendo 2 ou 3 para rodízio)
- Terminal (para gerar Base64)

---

### Passo 1: Fork do Repositório

1. Acesse → [https://github.com/CJackHwang/ds2api](https://github.com/CJackHwang/ds2api)
2. Clique no botão **Fork** (canto superior direito)
3. Escolha sua conta pessoal e faça o fork

---

### Passo 2: Preparar o config.json (Localmente)

```bash
git clone https://github.com/SEU-USUARIO/ds2api.git
cd ds2api
cp config.example.json config.json
```

Abra o `config.json` no Zed ou VS Code e configure:

```json
{
  "keys": ["sk-sua-chave-zed-2026-super-secreta"],
  "accounts": [
    {
      "email": "conta1@gmail.com",
      "password": "sua-senha-real",
      "token": ""
    },
    {
      "email": "conta2@gmail.com",
      "password": "sua-senha-real",
      "token": ""
    }
  ],
  "model_aliases": {
    "gpt-4o": "deepseek-chat",
    "claude-4": "deepseek-reasoner",
    "o3": "deepseek-reasoner",
    "gpt-5": "deepseek-reasoner"
  },
  "runtime": {
    "account_max_inflight": 3,
    "account_max_queue": 10
  }
}
```

---

### Passo 3: Gerar o Base64 (DS2API_CONFIG_JSON)

**Mac / Linux:**
```bash
base64 < config.json | tr -d '\n'
```

**Windows (PowerShell):**
```powershell
$json = Get-Content -Raw -Encoding UTF8 -Path "config.json" | Out-String
$bytes = [System.Text.Encoding]::UTF8.GetBytes($json.Trim())
$base64 = [Convert]::ToBase64String($bytes)
$base64 | Set-Clipboard
Write-Host "✅ Base64 copiado para a área de transferência! Cole direto no Vercel." -ForegroundColor Green
```

Copie **toda a linha gigante** gerada (Ctrl+C).

---

### Passo 4: Criar o Projeto no Vercel

1. Acesse [https://vercel.com/new](https://vercel.com/new)
2. Clique em **Import** → selecione seu fork (`SEU-USUARIO/ds2api`)
3. **Não clique em Deploy ainda** → clique em **"Configure Project"** ou em "Environment Variables"

---

### Passo 5: Configurar Environment Variables

Adicione estas variáveis (clique em "Add" para cada uma):

| Variável                        | Valor                                      | Tipo      |
|--------------------------------|--------------------------------------------|-----------|
| `DS2API_ADMIN_KEY`             | `senha-muito-forte-2026!` (crie uma forte) | Obrigatório |
| `DS2API_CONFIG_JSON`           | Cole aqui o Base64 inteiro                 | Altamente recomendado |
| `DS2API_VERCEL_INTERNAL_SECRET`| Mesma senha do ADMIN_KEY                   | Recomendado |

Opcional (para sincronização futura):
- `VERCEL_TOKEN` → crie em vercel.com/account/tokens
- `VERCEL_PROJECT_ID` → copie do projeto (Settings → General)
- `VERCEL_TEAM_ID` → se for team

Depois de adicionar todas, clique em **Deploy**.

---

### Passo 6: Aguardar o Deploy

- Tempo médio: 2 a 4 minutos (Vercel compila o WebUI React automaticamente)
- Quando aparecer **"Ready"** em verde, clique no link do seu domínio:
  - Exemplo: `https://ds2api-seunome.vercel.app`

---

### Passo 7: Testar Tudo

1. **Teste básico da API**  
   Abra no navegador:  
   `https://seu-domínio.vercel.app/v1/models`  
   Deve mostrar JSON com os modelos do DeepSeek.

2. **Acessar o Painel Admin**  
   `https://seu-domínio.vercel.app/admin`  
   - Usuário: `admin`  
   - Senha: a que você colocou em `DS2API_ADMIN_KEY`

3. **Teste de streaming** (importante)  
   Use no Zed ou Continue.dev e veja se o texto vem fluindo em tempo real.

---

### Passo 8: Configurar no Zed e VS Code (Continue.dev)

**Base URL:**  
`https://seu-domínio.vercel.app/v1`

Use exatamente a mesma `apiKey` que você colocou em `"keys"` no config.json.

Exemplo no settings do Zed:
```json
"api_url": "https://seu-domínio.vercel.app/v1"
```

---

### Como Atualizar a Configuração Depois (Sem Redeploy)

1. Acesse o painel `/admin`
2. Vá em **Settings** → edite o que quiser
3. Clique em **"Sync to Vercel"** (se você configurou VERCEL_TOKEN + PROJECT_ID)
   → O Vercel atualiza automaticamente a variável `DS2API_CONFIG_JSON`

---

### Dicas de Segurança & Melhores Práticas

- Use uma senha **muito forte** no `DS2API_ADMIN_KEY`
- Nunca compartilhe o link do Admin publicamente
- Se quiser mais proteção: ative **Vercel Deployment Protection** + Custom Domain
- Recomendado: 2~4 contas DeepSeek para nunca bater limite

---

### Troubleshooting (Problemas Comuns)

| Problema                          | Solução |
|-----------------------------------|--------|
| Build falhando (Go)               | No Vercel → Redeploy → "Clear Cache and Redeploy" |
| Streaming lento ou quebrado       | Certifique-se que `api/chat-stream.js` existe no repo |
| Página de "Authentication Required" | Project Settings → Security → desative Deployment Protection |
| 401 Unauthorized                  | Chave no Zed/Continue diferente da do config |
| WebUI em branco                   | Espere o primeiro deploy terminar (compila React) |

---

**Pronto!**  
Você agora tem um DS2API **público, seguro e sempre online** com DeepSeek Reasoner rodando em qualquer ferramenta.

Seu domínio final vai ser algo como `https://ds2api-seunome.vercel.app/v1`

