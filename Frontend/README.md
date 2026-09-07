# Guia de Logística — Frontend React

Frontend do Guia de Logística do CDD Campos dos Goytacazes, construído com
React e Vite. A aplicação é autenticada e consome a API Go do diretório
`Backend/`.

## Rodando

```bash
npm install
cp .env.example .env    # ajuste VITE_API_URL se a API Go não estiver em localhost:8080
npm run dev
```

Abre em `http://localhost:5173`.

## Como está organizado

```
src/
  api/ruas.js           # chamadas para a API Go (Backend/) — cai para dados de exemplo se a API não responder
  data/distritos.js      # cores e layout dos 9 distritos (mesmas cores da legenda atual)
  data/mockRuas.js        # dados de exemplo, no formato do model Rua do backend
  components/
    Header.jsx            # topo azul com busca e papel do usuário
    DistrictNav.jsx        # faixa amarela com os botões 601-609
    Sidebar.jsx            # navegação e identificação da unidade operacional
    Footer.jsx             # versão, contatos e instalação PWA no celular
    icons/Icons.jsx        # ícones SVG reutilizáveis, sem emojis
    MapPanel.jsx           # mapa principal com legenda e camada OSM
    LeafletMap.jsx         # mapa geográfico real
    AjustesRotasPanel/     # assistente de 3 passos: selecionar ruas → escolher distrito → confirmar
    OrdenamentoPanel/      # entrada manual, voz, câmera e sequência de entregas
    RuasTable.jsx          # tabela com busca por nome, CEP e distrito
  utils/buscaRua.js      # normalização, filtro e ranqueamento de ruas/CEPs
  App.jsx                  # junta tudo
```

## Integrações principais com o backend Go

- Autenticação JWT para os papéis `admin` e `colaborador`.
- Consulta de ruas, distritos, estatísticas, colaboradores, folgas e histórico.
- Ordenamento com entrada por texto, voz e scanner, resolução por nome/CEP,
  coordenadas e salvamento da sequência manual.
- APIs de redistritamento e ajustes de rotas protegidas para administradores.

## Comportamentos de interface

- A busca ignora acentos, pontuação, tipos de logradouro e artigos opcionais;
  prefixos e CEPs formatados também são aceitos. Quando uma palavra completa
  já existe entre as candidatas, resultados que só repetem esse trecho dentro
  de uma palavra maior são descartados.
- O ordenamento começa no CDD e escolhe a rua mais próxima da posição anterior
  em cada passo, medindo a menor distância até os pontos reais do traçado
  quando disponíveis, sem reordenação global posterior.
- O cabeçalho usa escudo para administradores e identificação para
  colaboradores.
- A área de instalação do PWA só aparece em layout móvel enquanto o app não
  estiver instalado. Sem prompt nativo, ela orienta a instalação pelo menu do
  navegador.
- O rodapé e a lateral usam a identidade visual do CDD sem emojis.

## Próximos passos sugeridos

1. Validar a busca com a lista operacional completa da unidade.
2. Revisar as 73 ruas sem geometria do snapshot local e as geometrias do OSM
   que reúnem trechos desconectados antes de considerar a coordenada definitiva.
3. Comparar a sequência sugerida com uma sequência operacional real; o cadastro
   atual não possui `rota` preenchida, então o ajuste do algoritmo depende dessa
   referência antes da publicação em produção. A planilha está pendente e será
   incorporada quando estiver disponível.
