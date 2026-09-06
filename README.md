# Guia de Logística — CDD Campos dos Goytacazes

Ferramenta interna da unidade CDD Campos dos Goytacazes (Correios): mapa interativo dos distritos postais, consulta de ruas/CEP, cadastro de colaboradores, ajuste de rotas e redistritamento.

Em produção: **https://cddcampos.devsaulo.com.br** — as alterações locais da
versão **v1.3.0** ainda aguardam publicação manual pelo workflow de deploy.

## Stack

- **Frontend**: React 19 + Vite, mapa real com React Leaflet + OpenStreetMap
- **Backend**: Go (Fiber + GORM), PostgreSQL 17 (PostGIS)
- **Autenticação**: JWT + bcrypt, com 2 papéis de acesso (`colaborador` e `admin`)
- **PWA**: instalável no celular, com Service Worker (cache somente de arquivos estáticos, nunca dados da API)
- **Deploy**: Docker Compose (postgres + api + frontend), atrás de Cloudflare Tunnel
- **CI/CD**: GitHub Actions com build-check antes do deploy e publicação manual via `workflow_dispatch`

## Estrutura

```text
Backend/
  cmd/              # comandos standalone: seed-admin, criar-usuario
  config/           # variáveis de ambiente
  database/         # conexão + migrations (GORM AutoMigrate)
  handlers/         # camada HTTP (Fiber)
  services/         # regras de negócio
  repositories/     # acesso ao banco
  models/           # structs GORM
  middlewares/      # autenticação JWT e tratamento de erros
  routes/           # definição das rotas e níveis de acesso

Frontend/
  src/
    api/             # chamadas para o backend (client.js centraliza token + erros)
    context/         # AuthContext (sessão, login/logout)
    components/      # UI (Header, Sidebar, MapPanel/LeafletMap, RuasTable, modais...)
  public/            # favicon, ícones PWA, manifest.json, sw.js
```

## Modelo de acesso

Todo o conteúdo operacional do sistema exige autenticação.

| Nível | O que faz | Precisa de login? |
|---|---|---|
| Infraestrutura | `/health` e `/swagger` | Não |
| Colaborador | Consultar mapa, ruas, CEP, distritos, aniversariantes, folgas, observações, Zé Rota e demais áreas liberadas ao colaborador | Sim |
| Admin | Todas as funções de colaborador + criar/editar/excluir dados, gerenciar usuários, ajustar rotas e executar redistritamento | Sim |

O sistema possui apenas **dois papéis de usuário**:

- `colaborador`
- `admin`

O login é feito com **matrícula + senha**.

O JWT possui validade de **8 horas**.

A conta é bloqueada temporariamente por **15 minutos após 5 tentativas incorretas**, além de limite de **10 tentativas por minuto por IP**.

As contas dos colaboradores são criadas inicialmente com senha provisória e o usuário é obrigado a trocar a senha no primeiro acesso.

## Rodando localmente

### Backend

```bash
cd Backend
cp .env.example .env
# ajuste DB_PASSWORD e gere um JWT_SECRET novo

go mod tidy
go run main.go
```

O backend sobe em:

```text
http://localhost:8080
```

As tabelas são criadas automaticamente pelo GORM AutoMigrate.

### Criar o primeiro admin

```bash
go run cmd/seed-admin/main.go -matricula SUA_MATRICULA -senha "SenhaTemporaria123"
```

### Frontend

```bash
cd Frontend
cp .env.example .env
# VITE_API_URL=http://localhost:8080

npm install
npm run dev
```

O frontend abre em:

```text
http://localhost:5173
```

## Deploy em produção

O deploy de produção é feito pelo **GitHub Actions**.

O `push` para o repositório não publica automaticamente em produção.

O workflow possui uma etapa de **build-check**, que:

1. Compila o backend em Go.
2. Executa o build do frontend React.
3. Só permite o deploy se ambas as etapas forem concluídas com sucesso.

Para publicar:

1. Acesse o repositório no GitHub.
2. Abra a aba **Actions**.
3. Selecione **Deploy Produção (VPS)**.
4. Clique em **Run workflow**.

Os secrets utilizados pelo workflow são:

- `VPS_HOST`
- `VPS_USER`
- `VPS_SSH_KEY`

Para informações detalhadas sobre produção, infraestrutura, procedimentos e pendências, consulte:

```text
nota-de-status-site-correios.md
```

## Funcionalidades

### Acesso restrito

Todo o conteúdo operacional exige autenticação.

Apenas os endpoints de infraestrutura `/health` e `/swagger` permanecem públicos.

### Mapa

- React Leaflet + OpenStreetMap
- Exibição dos distritos ativos
- Camada opcional com geometria real das ruas
- Suporte a geometria em linha e ponto
- Ferramenta para desenho manual de ruas sem geometria

### Consulta de ruas

Permite busca por:

- nome da rua
- CEP
- distrito

Também possui:

- exportação CSV
- impressão
- consulta por distrito

### Ordenamento de entregas

Área autenticada para preparar a carga do CDD Campos dos Goytacazes.

- ponto de partida fixo na unidade, com referência visual de unidade postal;
- entrada manual, por voz e leitura de etiqueta pela câmera;
- conversão de CEP para a rua correspondente no cadastro da unidade;
- busca tolerante a acentos, pontuação, abreviações e prefixos de digitação;
- agrupamento por rua e indicação clara de pendências ou coordenadas ausentes;
- sequência sugerida por proximidade, começando no CDD;
- cada parada escolhida pela menor distância à posição anterior, sem
  reordenação global posterior;
- ajuste manual da sequência pelo carteiro;
- histórico de correções e sequência habitual opcional, isolada por usuário;
- botão para limpar a carga atual sem apagar o histórico, reiniciando a data e
  hora da nova carga;
- disponível para `admin` e `colaborador`, respeitando as permissões de cada
  papel.

O sistema não usa aprendizado de máquina nesta etapa. As sequências habituais
continuam pessoais e só são reutilizadas quando a composição da carga coincide
com segurança.

### Redistritamento

Ferramenta administrativa para alteração da quantidade de distritos.

O fluxo de **redução** está implementado.

Exemplo:

```text
24 distritos → 17 distritos
```

O corte ocorre globalmente, sempre dos maiores códigos para os menores.

As ruas dos distritos que serão desativados ficam disponíveis para realocação manual em distritos sobreviventes.

O processo possui duas etapas:

- **Concluir** — salva o planejamento e permite continuar revisando.
- **Aplicar** — efetiva definitivamente o redistritamento.

Ao aplicar:

- `ruas.distrito` é atualizado
- distritos removidos recebem `ativo = false`
- nenhuma informação histórica é apagada
- o processo é registrado para auditoria

Também existe a opção **Voltar ao início**, enquanto o plano ainda não foi aplicado.

O sistema foi testado em produção com a remoção do distrito **624**, envolvendo a realocação de **120 ruas**.

O fluxo de **aumento de distritos** ainda não foi implementado na interface e na camada de negócio, embora a estrutura do banco já esteja preparada.

### Ajustes de Rotas

Permite que administradores movam ruas entre distritos.

Cada alteração mantém histórico com informações como:

- quem realizou a alteração
- data/hora
- distrito de origem
- distrito de destino

### Colaboradores

- cadastro
- consulta
- busca
- exclusão administrativa
- aniversariantes do dia

Foram criadas contas de acesso para os colaboradores utilizando a matrícula como login.

### Consulta de Folgas

Controle de saldo por matrícula utilizando um modelo de livro-razão.

Permite:

- consultar saldo
- lançar crédito/débito
- excluir lançamento

Alterações são restritas a administradores e possuem auditoria.

### Observações de rua

Permite registrar conhecimento operacional sobre as ruas, como:

- acesso difícil
- numeração fora de ordem
- múltiplos nomes
- características relevantes para entrega
- informações de segurança

O cadastro dessas informações é restrito a administradores.

### Zé Rota

Assistente interno disponível por texto e voz.

O Zé Rota utiliza o cadastro real do sistema para consultar:

- ruas
- distritos
- CEPs

Ele também consulta informações meteorológicas de Campos dos Goytacazes utilizando Open-Meteo.

Quando um endereço não é encontrado no cadastro interno, pode sugerir uma consulta externa pelo Google Maps, deixando claro que o resultado não faz parte da base oficial do sistema.

O assistente não deve inventar endereços inexistentes no cadastro.

### Gerenciamento de usuários

Administradores podem pela própria interface:

- criar contas
- definir papel
- resetar senha
- gerenciar usuários

Não é necessário acessar o banco ou terminal para operações comuns.

### PWA

O sistema pode ser instalado no celular.

O rodapé exibe **Instalar app** somente em layout móvel enquanto o aplicativo
não estiver instalado. Se o navegador não oferecer o prompt nativo, o botão
orienta o usuário a usar o menu de instalação do próprio navegador. Depois da
instalação, o botão é substituído por **App instalado** e permanece oculto nas
aberturas seguintes.

O Service Worker mantém apenas o casco estático da aplicação em cache.

Dados das APIs não são cacheados.

### Identidade visual

Os ícones da interface usam SVG de linha, sem emojis. O cabeçalho diferencia
os papéis com escudo para administradores e identificação de usuário para
colaboradores. O ponto de partida exibe um símbolo de unidade postal e a
lateral apresenta a marca Correios com a identificação operacional do CDD.

## Geometria das ruas

O banco possui suporte a geometria das ruas para melhorar a precisão do mapa.

Foram utilizados:

- OpenStreetMap
- Nominatim
- Overpass API

Parte das ruas foi automaticamente associada a geometrias reais.

O snapshot local consultado em 06/09/2026 tem **305 ruas sem geometria**, que
precisam ser desenhadas manualmente. A contagem pode mudar quando a base for
sincronizada novamente; o registro anterior de 303 ruas era de outro snapshot.

Resultados automáticos de baixa confiança foram descartados para evitar associação de ruas incorretas.

## Pendências conhecidas

- **305 ruas sem geometria real no snapshot local de 06/09/2026** — desenho
  manual em andamento; conferir novamente após sincronizar a base.
- **Fluxo de aumento do Redistritamento** — estrutura do banco pronta; faltam lógica de negócio e interface.
- **Ordenamento por rua** — futura tabela com sequência real de entrega (`rua_id`, `ordem`, `numero`).
- **Precisão do ordenamento** — as geometrias existentes foram auditadas e estão
  válidas; a primeira parada é fixada pela proximidade do CDD, mas falta uma
  lista operacional real para medir a sequência completa e orientar ajustes
  adicionais antes da publicação. A planilha está pendente e deve ser obtida
  nos próximos dois dias.
- **Zé Rota — próxima fase** — sugestão de rota para múltiplas encomendas, dependente do ordenamento das ruas.
- **AGC (Agência Comunitária)** — áreas sem entrega domiciliária ainda não modeladas.
- **Botão "Contribuir"** — planejado para alimentar o Zé Rota com características dos distritos.
- Resetar ou bloquear usuário ainda não invalida imediatamente um JWT já emitido; ele permanece válido até expirar em 8 horas.
- Botão **Por Carteiro** da tabela de ruas atualmente reordena os registros, mas não realiza um filtro real.

## Scripts auxiliares

Scripts Python utilizados para geometria:

```text
preencher_geometria_nominatim.py
upgradar_tracado_nominatim.py
casar_ruas_overpass.py
```

Eles são executados localmente e geram resultados para revisão humana antes de alterações no banco de produção.

Também existe:

```text
scripts/criar_logins_colaboradores.sql
```

Utilizado para criação em lote das contas dos colaboradores.

## Atualização e publicação

Fluxo normal no PC:

```bash
git add .
git commit -m "descricao da alteracao"
git push
```

Depois do `push`, a publicação em produção é disparada manualmente pelo GitHub Actions.

Isso permite manter o código versionado sem publicar automaticamente alterações para os colaboradores que utilizam o sistema em produção.
