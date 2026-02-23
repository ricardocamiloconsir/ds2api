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

Qualquer dúvida durante a instalação, cola o erro aqui que resolvemos em 2 minutos.
