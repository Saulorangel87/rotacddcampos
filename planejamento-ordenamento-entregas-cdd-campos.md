# NOVA FEATURE — ORDENAMENTO DE ENTREGAS
## Guia de Logística — CDD Campos dos Goytacazes

Quero implementar uma nova funcionalidade chamada **Ordenamento** no projeto existente.

Antes de alterar qualquer arquivo:

1. Analise a arquitetura atual do projeto.
2. Identifique os componentes existentes que podem ser reaproveitados.
3. Identifique models, services, repositories, handlers, routes e componentes React relacionados.
4. Preserve os padrões arquiteturais atuais.
5. Não refatore partes não relacionadas.
6. Não faça mudanças destrutivas no banco.
7. Não altere funcionalidades que já estão funcionando.
8. Trabalhe em etapas pequenas.
9. Rode build/testes a cada etapa relevante.
10. Se encontrar alguma decisão que possa alterar significativamente a arquitetura existente, pare e me explique antes de implementar.

O sistema está em produção e é utilizado por colaboradores reais.

O deploy de produção é manual via GitHub Actions. Portanto, trabalhar primeiro localmente e não assumir que qualquer alteração deve ser publicada automaticamente.

---

# 1. OBJETIVO DA FEATURE

A principal dor operacional que quero resolver é o tempo gasto pelo carteiro para organizar as encomendas antes de sair para entrega.

Um colaborador pode receber, por exemplo:

- 20 encomendas;
- 25 encomendas;
- 40 encomendas;
- 50 encomendas;
- ou mais.

Essas encomendas NÃO necessariamente pertencem ao mesmo distrito postal.

Um único colaborador pode sair com objetos distribuídos por vários distritos.

Portanto:

**NÃO pedir distrito para iniciar um ordenamento.**

O objetivo inicial NÃO é criar um GPS completo nem calcular a rota viária perfeita.

O objetivo do MVP é mais simples e prático:

> receber várias encomendas, identificar as ruas envolvidas e sugerir uma sequência lógica de ruas a partir do CDD Campos, utilizando proximidade geográfica.

A saída principal será uma **lista ordenada de ruas**.

Exemplo:

CDD Campos

1. Avenida Sete de Setembro
2. Rua Tenente Coronel Cardoso
3. Avenida Pelinca
4. Rua Voluntários da Pátria
5. Rua Barão de Miracema
6. ...

Neste primeiro momento, saber a ordem aproximada das ruas já resolve grande parte da dor operacional.

---

# 2. PONTO DE PARTIDA FIXO

Todo ordenamento deve começar no:

**CDD Campos dos Goytacazes**  
**Av. Sete de Setembro, 342**  
**Campos dos Goytacazes - RJ**

Este deve ser o ponto inicial fixo do algoritmo.

Não quero perguntar ao colaborador de onde ele está saindo nesta primeira versão.

O sistema deve possuir as coordenadas do CDD cadastradas/configuradas uma única vez.

Conceitualmente:

CDD = nó 0

Depois:

CDD
→ rua 1
→ rua 2
→ rua 3
→ ...
→ rua N

Não precisamos calcular retorno ao CDD nesta primeira versão.

---

# 3. ÁREA DE ATUAÇÃO

A feature deve trabalhar apenas com endereços da região de:

**Campos dos Goytacazes - RJ - Brasil**

Não limitar por distrito postal.

A rota pode atravessar vários distritos.

Ao realizar qualquer geocodificação externa, sempre contextualizar a busca com:

- Campos dos Goytacazes
- Rio de Janeiro
- Brasil

Evitar aceitar resultados claramente localizados em outra cidade.

Não criar um bloqueio rígido baseado nos polígonos dos distritos, porque esta feature deve poder trabalhar com múltiplos distritos em um mesmo ordenamento.

---

# 4. ESCOPO DO MVP

Neste primeiro momento, a unidade principal de ordenação será a:

## RUA

Não precisamos calcular ainda a posição exata de cada número do logradouro.

Exemplo:

Encomendas:

- Avenida Pelinca, 120
- Avenida Pelinca, 450
- Avenida Pelinca, 700

Para o algoritmo inicial podemos considerar:

Avenida Pelinca  
3 objetos

Ou seja:

50 encomendas
↓
30 ruas diferentes
↓
ordenar as 30 ruas

O número do endereço deve ser preservado nos dados do objeto sempre que estiver disponível, porque será útil futuramente.

Mas ele NÃO deve ser requisito para gerar o ordenamento nesta primeira versão.

---

# 5. TRÊS FORMAS DE ENTRADA

Quero permitir que o colaborador adicione encomendas de três maneiras:

1. Digitação
2. Voz
3. Scanner da etiqueta/código

Todas as três formas de entrada devem convergir para o mesmo fluxo interno.

Arquitetura conceitual:

DIGITAÇÃO ─┐
           │
VOZ ───────┼──→ normalização → identificação da rua → objeto
           │
SCANNER ───┘

Não criar três sistemas separados.

Criar uma camada comum de normalização/resolução.

---

# 6. ENTRADA MANUAL

O colaborador poderá escrever algo como:

"Pelinca 520"

ou:

"Avenida Pelinca 520"

ou:

"Rua Tenente Coronel Cardoso 300"

O sistema deve tentar identificar a rua utilizando primeiro o banco interno do CDD.

Prioridade:

1. buscar no cadastro interno;
2. normalizar texto;
3. tentar encontrar correspondência;
4. usar geocodificação externa apenas quando necessário.

O banco interno deve continuar sendo a principal referência para os nomes das ruas.

Evitar cadastrar livremente uma rua diferente quando já existir uma correspondente no banco.

---

# 7. ENTRADA POR VOZ

A entrada por voz deve ser apenas uma alternativa à digitação.

Exemplo:

Usuário fala:

"Rua Tenente Coronel Cardoso número quatrocentos"

Transformar em texto.

Depois utilizar exatamente o mesmo pipeline da entrada manual.

Fluxo:

voz
→ texto
→ normalização
→ busca da rua
→ criação do objeto

O projeto já possui recursos de voz no Zé Rota.

Antes de criar uma implementação nova do zero:

- inspecionar o que já existe;
- reaproveitar a infraestrutura quando fizer sentido;
- evitar duplicação de lógica.

---

# 8. SCANNER

Quero um botão de:

**Escanear encomenda**

e não simplesmente "Ler código de rastreamento".

Objetivo:

usar a câmera do celular para tentar extrair dados úteis da etiqueta da encomenda.

A implementação deve ser projetada de maneira flexível.

Pode haver diferentes códigos na etiqueta.

Não assumir que todo código lido obrigatoriamente contém endereço completo.

O scanner deve:

1. detectar o código;
2. identificar o conteúdo retornado;
3. extrair os campos úteis que estiverem disponíveis;
4. tentar associar o objeto a uma rua cadastrada;
5. se não houver informação suficiente para identificar a rua, deixar o objeto pendente para complementação manual ou por voz.

Se for possível extrair:

- código de rastreamento;
- CEP;
- número;
- complemento;
- coordenadas;
- outros dados úteis;

preservar apenas o que for necessário para a funcionalidade.

Não armazenar dados pessoais desnecessários do destinatário.

---

# 9. PRIVACIDADE

Este sistema trabalha com dados operacionais reais.

Para a funcionalidade de ordenamento, armazenar apenas o necessário.

Priorizar:

- código do objeto, quando necessário;
- rua;
- número;
- complemento, quando necessário;
- CEP;
- latitude;
- longitude;
- origem da informação.

Evitar armazenar:

- nome do destinatário;
- CPF;
- telefone;
- outros dados pessoais sem necessidade operacional.

Se o scanner retornar dados extras, descartá-los se não forem necessários para o ordenamento.

---

# 10. AGRUPAMENTO DAS ENCOMENDAS

Várias encomendas podem pertencer à mesma rua.

Exemplo:

Objeto A → Avenida Pelinca  
Objeto B → Avenida Pelinca  
Objeto C → Rua Voluntários  
Objeto D → Avenida Pelinca

Resultado:

Avenida Pelinca  
3 objetos

Rua Voluntários  
1 objeto

Portanto, diferenciar:

OBJETO

de

RUA/PARADA DE ORDENAÇÃO

Neste MVP, a lista é organizada principalmente por ruas.

Não criar várias posições consecutivas para a mesma rua se puderem ser agrupadas.

---

# 11. MODELO CONCEITUAL

Não implemente obrigatoriamente estes nomes sem antes analisar os models existentes.

Use isso como referência conceitual.

Podemos ter algo equivalente a:

Ordenamento
- id
- usuario_id
- status
- criado_em
- atualizado_em
- ordem_gerada
- ordem_final

ObjetoOrdenamento
- id
- ordenamento_id
- codigo_rastreamento
- rua_id
- numero
- complemento
- cep
- latitude
- longitude
- origem_entrada
- status_resolucao

RuaOrdenamento / Parada
- rua_id
- nome_rua
- latitude_referencia
- longitude_referencia
- quantidade_objetos
- ordem_sugerida
- ordem_final

Não duplicar desnecessariamente informações que já existem na tabela `ruas`.

Preferir relacionamento por `rua_id`.

---

# 12. COORDENADAS

O banco já possui geometria para várias ruas, mas nem todas as ruas possuem geometria.

Portanto:

NÃO depender exclusivamente da geometria existente.

Fluxo desejado:

Rua encontrada
↓
tem coordenada/geometria utilizável no banco?
↓
SIM
→ usar

NÃO
→ tentar resolver externamente
→ validar que pertence a Campos dos Goytacazes
→ utilizar uma posição aproximada da rua
→ considerar armazenar/cachear o resultado para não consultar novamente no futuro

O objetivo é obter uma coordenada representativa da rua.

Não precisamos inicialmente localizar exatamente o número 520 dentro da rua.

Uma coordenada aproximada da rua já é suficiente para a primeira versão.

---

# 13. GEOCODIFICAÇÃO EXTERNA

Nesta primeira versão quero evitar APIs pagas.

Portanto, não integrar Google Maps / Routes API / Route Optimization API agora.

Podemos utilizar alternativas gratuitas/open source quando necessário, respeitando limites e políticas de uso.

O sistema já utilizou OpenStreetMap/Nominatim anteriormente.

Antes de criar algo novo:

- verificar códigos existentes;
- verificar scripts/serviços relacionados;
- reaproveitar normalização já existente se aplicável.

Toda consulta deve conter contexto territorial:

nome da rua
+
Campos dos Goytacazes
+
Rio de Janeiro
+
Brasil

Exemplo:

"Rua X, Campos dos Goytacazes, Rio de Janeiro, Brasil"

Se o resultado vier de outra cidade, rejeitar.

Evitar consultas externas repetidas para a mesma rua.

Criar cache/persistência quando apropriado.

---

# 14. ALGORITMO DE ORDENAÇÃO

Nesta primeira versão quero um algoritmo gratuito executado no nosso próprio backend.

Não usar IA.

Não usar Google Routes.

Não usar serviço pago.

Começar pelo ponto fixo do CDD.

Usar as coordenadas das ruas para criar uma sequência aproximada.

Primeira heurística sugerida:

## Nearest Neighbor

Fluxo:

1. posição atual = CDD;
2. procurar a rua restante geograficamente mais próxima;
3. adicioná-la à sequência;
4. posição atual = rua escolhida;
5. repetir até terminar.

Para distância entre coordenadas, utilizar cálculo geográfico apropriado, como Haversine.

Quando a rua possuir uma geometria com vários vértices, a distância da etapa
é a menor Haversine entre a posição atual e os vértices disponíveis. Depois da
escolha, a posição atual passa para o vértice que produziu essa menor distância;
a coordenada média fica apenas como fallback para exibição.

Nesta primeira versão, não aplicar 2-opt nem outra melhoria global. Essas
técnicas podem trocar uma rua que é a mais próxima da posição atual por uma
sequência globalmente menor, contrariando o critério operacional definido para
o MVP. A sequência deve permanecer exatamente na ordem das escolhas locais:

ruas
↓
Nearest Neighbor a partir do CDD
↓
ordem sugerida

Qualquer melhoria global fica postergada até existir uma referência operacional
real que permita medir o resultado e aprovar o novo critério.

Importante:

Esta solução NÃO conhece necessariamente o percurso real pelas vias.

Portanto, não apresentar ao usuário como:

"rota mais rápida"

ou

"melhor rota possível".

Usar linguagem como:

**Ordem sugerida**

ou:

**Ordenamento por proximidade**

---

# 15. RESULTADO

Após o cadastro dos objetos:

Exemplo:

47 objetos  
22 ruas

Botão:

**Gerar ordenamento**

Resultado:

ORDENAMENTO SUGERIDO

Partida:  
CDD Campos  
Av. Sete de Setembro, 342

1. Avenida Sete de Setembro  
   2 objetos

2. Rua Tenente Coronel Cardoso  
   4 objetos

3. Avenida Pelinca  
   3 objetos

4. Rua Voluntários da Pátria  
   1 objeto

5. Rua Barão de Miracema  
   5 objetos

...

Não precisamos inicialmente criar navegação curva a curva.

Não precisamos gerar instruções:

- vire à esquerda;
- vire à direita;
- siga 500 metros.

O produto inicial é a LISTA ORDENADA.

---

# 16. CORREÇÃO MANUAL DA ORDEM

O carteiro conhece a área e deve poder corrigir a sugestão.

Permitir reordenar as ruas manualmente.

Preferencialmente usando drag-and-drop.

Exemplo sugerido:

1. Rua A
2. Rua B
3. Rua C
4. Rua D

Carteiro muda para:

1. Rua A
2. Rua C
3. Rua B
4. Rua D

Salvar:

- ordem sugerida pelo algoritmo;
- ordem final escolhida pelo colaborador.

Esses dados poderão ser úteis no futuro para melhorar o sistema.

Não implementar machine learning agora.

Apenas preservar o histórico quando fizer sentido.

### Evolução local — 06/09/2026: preservação e memória pessoal

Decisão: evoluir em etapas, mantendo as correções individuais separadas de
referências coletivas. Publicar em produção somente após concluir as etapas
e validar o conjunto com o responsável pela unidade.

Implementado nesta etapa local:

- “Gerar novamente” preserva a ordem final salva quando as ruas da carga são as mesmas;
- cada novo salvamento registra autor, data, sugestão original e sequência escolhida;
- o histórico de correções fica separado das paradas temporárias e sobrevive a “Limpar lista”;
- “Usar como minha sequência habitual” é opcional e vale apenas para o próprio usuário;
- a reutilização automática exige o mesmo conjunto de cadastros de ruas e a mesma origem no CDD;
- nomes de ruas, sozinhos, não autorizam reutilização: grupos de trechos sem cadastro específico continuam permitindo ajuste da carga, mas precisam ser identificados pelo CEP para a memória pessoal;
- recalcular somente pelo algoritmo preserva o histórico e a referência habitual para uso futuro;
- o histórico mostra os últimos 20 salvamentos, permite preparar uma versão anterior para a carga compatível e desativar uma sequência habitual;
- salvar a ordem e seu histórico ocorre em uma transação; cada sequência habitual possui uma única versão ativa por usuário, origem e conjunto de ruas.

Limites e próximos passos:

- salvamentos feitos antes desta etapa não possuem histórico retroativo; a ordem final ainda presente é preservada ao gerar novamente e pode ser salva como habitual;
- mudanças na composição da carga invalidam a lista calculada; a memória habitual é reaplicada quando o conjunto completo de cadastros volta a coincidir;
- ainda não adaptar automaticamente uma referência a listas parciais ou ruas novas;
- melhorar coordenadas e precisão do motor com exemplos concretos;
- planejar referências por área/trecho e revisão das sugestões pelo admin antes de qualquer influência sobre outros usuários;
- manter versões e reversão também na futura publicação coletiva; quantidade de salvamentos não deve equivaler a aprovação.

Validação: testes Go de preservação, isolamento, incompatibilidade e desativação;
teste de integração PostgreSQL opt-in em schema isolado com rollback; testes Node
de restauração por cadastro e build do frontend. O compartilhamento coletivo
continua pendente, sem aprendizado de máquina nesta fase.

### Evolução local — 06/09/2026: refinamento criterioso da busca

Esta melhoria permanece dentro da Fase 8 — Refinamento e aperfeiçoa o pipeline
de entrada das Fases 2 e 7, sem alterar as permissões ou o escopo do MVP:

- nomes são comparados sem acentos, pontuação, tipo do logradouro e artigos opcionais;
- todos os termos relevantes precisam corresponder, com prefixos controlados para digitação progressiva;
- correspondências exatas vencem resultados parciais; homônimos continuam pendentes para escolha explícita;
- CEPs são pesquisados com ou sem hífen e podem ser usados no cabeçalho e na tela de consulta;
- a busca visual evita escolher o primeiro resultado apenas por ordem alfabética;
- respostas antigas não são sobrescritas por consultas mais recentes na tela de CEP;
- testes cobrem Silva Tavares, abreviações, acentos, artigos omitidos, prefixos e CEP.

A validação inicial da busca foi concluída com o cadastro local. A próxima
parte da Fase 8 é obter uma referência operacional real e, com ela, medir a
precisão das coordenadas e do ordenamento antes da publicação conjunta em
produção.

### Evolução local — 06/09/2026: validação no cadastro sincronizado

Foi feita uma leitura somente no PostgreSQL local para conferir a base usada
pela busca, sem criar, editar ou excluir registros. O snapshot encontrado foi:

- 2.115 ruas cadastradas em 24 distritos;
- 1.810 ruas com geometria e 305 ainda sem geometria;
- `Silva Tavares`: 2 resultados, mantendo a rua homônima distinta de
  `Pedreval da Silva Tavares`;
- `Av. Sete Setembro`: os quatro trechos da avenida permanecem priorizados;
  o filtro da interface remove `Rua Vinte e Oito de Setembro`, que só coincidia
  pelo prefixo `Sete` dentro de `Setembro`;
- `sergio`: 3 opções para escolha explícita;
- `28010-562`: 1 trecho identificado diretamente pelo CEP;
- `Santa Cecilia`: 2 opções homônimas, sem escolher automaticamente uma delas.

O filtro compartilhado agora exige a palavra exata quando ela existe entre as
candidatas e mantém prefixo apenas para a parte ainda digitada. A validação
local passou de 8 para 10 testes de frontend, além do build Vite. A contagem
de geometrias é um retrato da base local nesta data; deve ser conferida de novo
quando a produção for sincronizada.

Próxima etapa da Fase 8: obter uma sequência operacional real de paradas ou
uma planilha de ordem por rua para comparar com a sugestão do motor.

### Evolução local — 06/09/2026: auditoria de qualidade para o ordenamento

Antes de alterar o motor, foi feita uma leitura somente do cadastro local para
separar problemas de dados de problemas de algoritmo:

- as 1.810 geometrias existentes são válidas, estão dentro dos limites amplos
  de Campos dos Goytacazes e não há traçados degenerados;
- a distribuição é de 1.685 `MultiLineString`, 114 `LineString` e 11 `Point`;
- há 305 ruas sem geometria, concentradas principalmente nos distritos 608
  (35), 611 (30), 606 (27), 623 (23), 615 (19) e 621 (16);
- o campo `rota` está vazio em todo o cadastro local, então ele não pode ser
  usado para ordenar ou priorizar trechos;
- o cache externo possui somente 2 resultados encontrados e 4 negativos, sem
  evidência de que uma coordenada inválida esteja sendo usada no motor.

Conclusão: não há evidência de defeito estrutural nas geometrias atuais. A
ordem gerada continua sendo uma sugestão por distância em linha reta, como
previsto no MVP. Para aprimorar a precisão com segurança, falta uma referência
operacional (por exemplo, a sequência real de paradas ou uma planilha de
ordem por rua). Essa referência será usada para medir o erro e ajustar o motor
sem transformar uma preferência arbitrária em regra geral.

Pendência registrada: a referência operacional deverá ser obtida em até dois
dias. Até que ela esteja disponível, não fazer novos ajustes baseados em
preferência de rota; seguir com as verificações independentes de acesso,
interface e preparação para produção.

Validação desta auditoria: consulta PostgreSQL local somente leitura, `go test
./...`, 10 testes Node e `npm run build`. Nenhuma tabela ou geometria foi
alterada.

### Evolução local — 06/09/2026: auditoria do algoritmo de proximidade

O motor foi revisado contra o contrato do MVP e agora aplica somente o vizinho
mais próximo:

- a posição inicial é o ponto fixo do CDD;
- em cada passo, escolhe a rua restante com menor distância Haversine à posição
  anterior;
- a entrada não é modificada e cada parada permanece uma única vez na saída;
- não há `2-opt` ou outra reordenação global que possa trocar uma escolha local
  mais próxima por uma sequência globalmente menor;
- um teste reproduz o caso do cadastro local em que a melhoria anterior trocava
  `RUA CORONEL WALTER KRAMER` e `RUA HUMBERTO DE CAMPOS`; a ordem estrita agora
  mantém a rua mais próxima em cada passo.

Com isso, a garantia do MVP fica explícita: cada transição parte da coordenada
da parada anterior e escolhe a menor distância geográfica disponível. Isso
ainda não representa o sentido real das vias nem a ordem dos números de
entrega; essa etapa continua dependente da referência operacional pendente.

### Verificação do caso Araújo/Advaldo — 06/09/2026

No teste com a carga real do ordenamento, a coordenada cadastrada para `RUA
ADVALDO MACIEL` ficou cerca de 244 m do centro de `RUA ARAÚJO SILVA` e cerca
de 113 m pelo ponto mais próximo das geometrias. A sequência estrita agora
coloca Advaldo logo após Araújo quando as duas ruas estão na mesma carga. O
caso anterior usava por engano `RUA ALCIDES VIEIRA MACIEL`, que pertence a outra
posição do cadastro e fica aproximadamente 3,03 km de Araújo. Nenhuma
coordenada foi alterada nesta etapa.

### Evolução local — 07/09/2026: auditoria dos pontos usados na proximidade

O laço do motor continua seguindo o contrato do MVP: começa no ponto fixo do
CDD Campos dos Goytacazes e, a cada passo, escolhe a rua restante com menor
distância Haversine à posição atual. A auditoria encontrou uma distorção na
representação das ruas: o resolvedor calculava um centro médio com todos os
vértices da geometria. O importador do OSM agrupa ways de um mesmo nome antes
de gravá-los, então alguns registros carregam trechos desconectados em bairros
distintos; nesses casos, o centro médio pode ficar fora do traçado.

O resolvedor passou a preservar os vértices internos. O otimizador calcula a
menor distância até qualquer ponto disponível da rua e atualiza a posição
atual para o ponto escolhido antes de procurar o próximo vizinho. A coordenada
média permanece somente como fallback de exibição e compatibilidade; uma
coordenada externa sem traçado continua sendo tratada como um único ponto.

No retrato local foram encontradas 2.042 geometrias válidas no formato, 73
ruas sem geometria e 207 geometrias cujo centro médio fica a mais de 500 m de
qualquer vértice. Essas 207 são candidatas a revisão do casamento OSM por
distrito/CEP; não foram sobrescritas automaticamente. A alteração reduz o
efeito dessa distorção no ordenamento, mas a confirmação operacional continua
necessária para vias repetidas ou trechos de bairros diferentes. A comparação
por ID dos hashes JSONB das 2.115 linhas ativas encontrou zero divergências
entre o banco local e a produção; a produção também mantém 73 ruas sem
geometria.

Testes determinísticos verificam a origem no CDD, a menor distância em todas as
transições e o uso de pontos reais quando o centro médio é enganoso. O backend
foi validado com `go test ./... -count=1` e o frontend com `npm run build`.

### Evolução local — 06/09/2026: identidade visual, ciclo de carga e instalação

Refinamento concluído dentro da Fase 8, sem alterar o modelo de permissões ou
as regras do motor de ordenamento:

- os emojis funcionais foram substituídos por ícones SVG de linha, mantendo
  proporção, contraste e rótulos acessíveis;
- o cabeçalho diferencia os papéis com escudo para `admin` e identificação de
  usuário para `colaborador`, além dos rótulos `ADMIN` e `COLAB.`;
- o ponto de partida passou a exibir um símbolo de unidade postal junto da
  sigla `CDD`, enquanto o título mostra `Campos dos Goytacazes`;
- a marca institucional da lateral foi reorganizada com o logotipo dos
  Correios, identificação da unidade operacional e descrição da função do
  painel;
- ao usar “Limpar lista”, a carga anterior é removida em transação e o horário
  do ordenamento é reiniciado para a nova carga; histórico e referências
  pessoais permanecem preservados;
- o instalador PWA aparece somente em layout móvel enquanto o app não está
  instalado; quando o prompt nativo não está disponível, o botão orienta o
  caminho manual do navegador; após a instalação, mostra “App instalado” e
  permanece oculto nas próximas aberturas;
- a detecção móvel acompanha redimensionamentos e o modo responsivo para não
  perder a opção durante testes em celular.
- a tela de novidades é exibida depois do login uma vez por usuário e por
  versão; a entrega atual foi identificada como `1.4.0` e inclui o ordenamento
  estrito por proximidade passo a passo.

Validação desta etapa: `go test ./... -count=1`, `npm run build`,
`git diff --check` e conferência visual local no navegador. A publicação em
produção continua pendente até concluir a validação operacional da busca,
coordenadas e precisão do ordenamento.

---

# 17. SIDEBAR — ORGANIZAÇÃO

A interface atual possui sidebar lateral.

Adicionar nova opção:

**Ordenamento**

NÃO usar:

"Novo Ordenamento"

como nome da opção da sidebar.

Motivo:

a sidebar representa áreas do sistema.

"Novo ordenamento" deve ser uma ação dentro da página.

A nova opção deve ficar:

**logo abaixo de "Mapa Geral".**

Organização desejada para usuários colaboradores:

Mapa Geral  
Ordenamento  
Ruas  
CEP  
Folgas  
Colaboradores  
Relatórios

As opções administrativas continuam condicionadas ao papel `admin`.

Hoje:

- Redistritamento → admin
- Ajustes de Rotas → admin

Manter essa regra.

Para administradores, organizar visualmente de forma coerente.

Sugestão:

Mapa Geral  
Ordenamento

[operações normais]  
Ruas  
CEP  
Folgas  
Colaboradores  
Relatórios

[administração]  
Redistritamento  
Ajustes de Rotas  
Usuários

Não necessariamente adicionar textos grandes de seção se isso prejudicar o layout atual.

Primeiro analisar como a sidebar foi construída.

O objetivo principal é:

- Ordenamento logo abaixo de Mapa Geral;
- funções operacionais fáceis de encontrar;
- funções administrativas claramente separadas;
- manter a identidade visual atual.

Não redesenhar toda a sidebar.

---

# 18. PERMISSÕES

A área de **Ordenamento** deve estar disponível para:

- colaborador
- admin

Não é uma ferramenta exclusiva de administração.

Redistritamento e Ajustes de Rotas continuam visíveis apenas para administradores conforme a regra atual do sistema.

Garantir proteção também no backend quando houver endpoints relacionados ao ordenamento.

Não confiar apenas em esconder menu no frontend.

---

# 19. TELA PRINCIPAL DE ORDENAMENTO

Ao clicar em:

Ordenamento

abrir página:

# Ordenamento de Entregas

Estado inicial:

Ponto de partida

CDD Campos  
Av. Sete de Setembro, 342

Nenhum ordenamento em andamento.

Botão:

+ Novo ordenamento

Ao iniciar:

Novo ordenamento

0 objetos  
0 ruas

Adicionar encomenda:

[ Escanear ]  
[ Falar ]  
[ Digitar ]

Lista abaixo:

Encomendas adicionadas

Quando houver objetos:

25 objetos  
18 ruas

Botão:

Gerar ordenamento

Não criar uma tela excessivamente carregada.

Priorizar uso no celular.

O carteiro provavelmente utilizará esta funcionalidade principalmente em dispositivo móvel.

---

# 20. FLUXO DE DIGITAÇÃO

Ao tocar em:

Digitar

mostrar campo simples.

Exemplo:

[ Avenida Pelinca 520 ]

Após confirmação:

- normalizar;
- identificar rua;
- associar ao cadastro;
- adicionar objeto;
- fechar/limpar campo para permitir cadastro rápido do próximo.

O fluxo precisa ser rápido porque o colaborador pode inserir dezenas de objetos.

---

# 21. FLUXO DE VOZ

Ao tocar:

Falar

usar o recurso de voz.

Mostrar claramente o texto reconhecido antes ou durante a confirmação.

Exemplo:

Reconhecido:  
"Avenida Pelinca 520"

[Adicionar]

Se houver integração de voz já implementada no Zé Rota, avaliar reaproveitamento.

---

# 22. FLUXO DO SCANNER

Ao tocar:

Escanear

abrir câmera.

Após leitura bem-sucedida:

feedback visual imediato.

Exemplo:

✓ Encomenda adicionada

Avenida Pelinca

O usuário deve conseguir continuar escaneando objetos rapidamente.

Evitar fluxo com muitos modais/confirmações se a identificação for confiável.

Quando não conseguir identificar a rua:

⚠ Não foi possível identificar a rua.

Opções:

[Completar digitando]  
[Falar endereço]  
[Cancelar]

---

# 23. OBJETOS PENDENTES

Criar conceito de objeto pendente/revisão.

Exemplo:

25 objetos

23 identificados  
2 precisam de revisão

Não gerar ordenamento utilizando silenciosamente dados duvidosos.

Apresentar claramente os objetos que precisam ser corrigidos.

---

# 24. ENDEREÇOS REPETIDOS

Normalizar os nomes antes de agrupar.

Exemplos como:

Av Pelinca  
Avenida Pelinca  
AV. PELINCA

devem ser associados à mesma rua cadastrada quando possível.

Não fazer agrupamento apenas por string crua.

Utilizar `rua_id` sempre que disponível.

---

# 25. INTERFACE MOBILE

Esta funcionalidade deve ser pensada mobile-first.

Os principais botões precisam ser grandes o suficiente para toque.

Prioridade visual:

1. Escanear
2. Falar
3. Digitar

ou outro arranjo que faça sentido após analisar o design atual.

O scanner provavelmente será o fluxo mais rápido no uso operacional.

Não prejudicar desktop.

---

# 26. MAPA

Não considerar mapa obrigatório para o MVP.

Primeira entrega:

entrada
→ processamento
→ ordenamento
→ lista

Se for simples aproveitar o mapa existente para mostrar pontos numerados, isso pode ser planejado para etapa posterior.

Não deixar o mapa atrasar o MVP.

---

# 27. HISTÓRICO

Não quero transformar o MVP em um sistema gigantesco.

Mas preparar a arquitetura para futuramente permitir:

- salvar ordenamentos;
- reabrir;
- consultar histórico;
- comparar ordem sugerida x ordem final.

Se implementar persistência agora for natural dentro da arquitetura existente, podemos armazenar.

Caso complique demasiadamente o MVP, explique antes.

---

# 28. NÃO IMPLEMENTAR AGORA

Não implementar nesta primeira fase:

- Google Maps Routes API;
- Google Route Optimization API;
- API paga;
- navegação curva a curva;
- GPS em tempo real;
- cálculo de trânsito;
- otimização pelo número exato da residência;
- lado par/ímpar da rua;
- ordem real dos números;
- machine learning;
- IA para calcular rota;
- previsão de tempo da entrega;
- múltiplos veículos;
- divisão automática de carga entre carteiros;
- retorno otimizado ao CDD;
- seleção obrigatória de distrito.

Esses pontos podem ser evoluções futuras.

---

# 29. EVOLUÇÕES FUTURAS

A arquitetura não deve impedir futuramente:

V2
- melhorar coordenadas;
- usar número;
- melhorar geocodificação;
- mapa com sequência numerada.

V3
- ordenamento dentro da própria rua;
- lado da rua;
- par/ímpar;
- sequência real dos números.

V4
- utilizar conhecimento operacional;
- observações das ruas;
- preferências;
- histórico de correções.

V5
- integrar um motor real de roteamento;
- Google;
- OSRM;
- GraphHopper;
- Valhalla;
- outro provider.

Por isso, isolar o motor de ordenamento atrás de uma interface/service.

Exemplo conceitual:

RouteOptimizer  
    Optimize(...)

Implementação inicial:

LocalProximityOptimizer

No futuro:

GoogleRouteOptimizer  
OSRMRouteOptimizer  
etc.

Não acoplar o frontend diretamente a qualquer provider externo.

---

# 30. BACKEND

Seguir a arquitetura já existente:

routes
→ handlers
→ services
→ repositories
→ database/models

Não colocar regra de negócio complexa nos handlers.

Criar um service específico para ordenamento.

Separar:

- resolução de endereço;
- coordenadas/geocodificação;
- algoritmo;
- persistência.

Não misturar tudo em uma única função gigante.

---

# 31. FRONTEND

Seguir os padrões existentes do projeto.

Antes de criar novos componentes:

- analisar Sidebar;
- AuthContext;
- client.js/apiFetch;
- páginas existentes;
- modais;
- componentes de voz;
- mapa;
- padrões CSS.

Não duplicar cliente HTTP.

Usar a autenticação existente.

Não criar um segundo sistema de autenticação.

---

# 32. PWA

O projeto é PWA.

Não alterar a política atual de cache das APIs.

Dados de ordenamento não devem ficar presos em cache de Service Worker como se fossem conteúdo estático.

Manter o comportamento atual de não cachear dados dinâmicos da API.

---

# 33. SEGURANÇA

Todos os endpoints da feature Ordenamento devem exigir autenticação.

Papéis permitidos:

- colaborador
- admin

Manter os padrões de middleware existentes.

Não expor dados de ordenamento em rotas públicas.

---

# 34. MIGRATIONS

O projeto usa GORM AutoMigrate.

Antes de adicionar tabelas/colunas:

1. analisar models atuais;
2. verificar padrões de migrations;
3. propor o menor conjunto possível de novas estruturas;
4. evitar alteração destrutiva;
5. preservar compatibilidade com produção.

---

# 35. TESTES

Criar testes principalmente para lógica pura.

Prioridade:

## Algoritmo

Testar:

- distância;
- nearest neighbor;
- escolha da rua mais próxima em cada passo;
- uma rua;
- duas ruas;
- dezenas de ruas;
- coordenadas duplicadas;
- entrada vazia;
- coordenadas inválidas.

## Agrupamento

Testar:

- várias encomendas da mesma rua;
- variações de nome;
- associação por `rua_id`.

## Validação territorial

Testar:

- resultado em Campos;
- resultado fora de Campos.

## Permissões

Testar:

- colaborador acessa Ordenamento;
- admin acessa Ordenamento;
- usuário não autenticado recebe 401.

---

# 36. LOGS

Adicionar logs úteis sem expor dados pessoais.

Pode registrar:

- ordenamento criado;
- quantidade de objetos;
- quantidade de ruas;
- falha de resolução;
- execução do algoritmo;
- duração.

Não logar dados pessoais desnecessários.

---

# 37. PERFORMANCE

O algoritmo local deve suportar tranquilamente casos como:

25 ruas  
50 ruas  
100 ruas

Não precisa otimização prematura.

Nearest Neighbor é suficiente inicialmente. O motor deve ser claro e testável,
sem trocar escolhas locais por uma melhoria global sem validação operacional.

---

# 38. UX — PRINCÍPIO CENTRAL

O processo deve exigir o menor número possível de interações.

O carteiro pode estar ordenando dezenas de encomendas.

Cada clique desnecessário multiplicado por 50 vira um problema.

Priorizar:

leu
→ identificou
→ adicionou
→ próximo

em vez de:

leu
→ abriu modal
→ confirmou
→ fechou modal
→ voltou
→ próximo

Confirmação manual deve aparecer principalmente quando houver dúvida.

---

# 39. NÃO QUEBRAR O SISTEMA ATUAL

Partes críticas existentes:

- autenticação;
- mapa;
- ruas;
- CEP;
- folgas;
- colaboradores;
- relatórios;
- usuários;
- Redistritamento;
- Ajustes de Rotas;
- Zé Rota.

A nova feature não deve alterar comportamento dessas áreas sem necessidade.

---

# 40. PRIMEIRA ETAPA DO TRABALHO

ANTES DE IMPLEMENTAR, quero que você faça uma análise do projeto e me entregue:

1. arquivos que serão alterados;
2. arquivos novos necessários;
3. componentes existentes que podem ser reaproveitados;
4. proposta de modelagem do banco;
5. endpoints necessários;
6. fluxo do frontend;
7. como resolveremos coordenadas;
8. como será implementado o algoritmo;
9. bibliotecas adicionais eventualmente necessárias;
10. riscos;
11. divisão da implementação em fases.

Não comece criando dezenas de arquivos sem primeiro entender o projeto.

Depois da análise, implementar fase por fase.

---

# 41. DIVISÃO SUGERIDA

## Fase 1 — Estrutura

- item Ordenamento na sidebar;
- rota frontend;
- página inicial;
- permissões;
- estrutura backend;
- models necessários.

## Fase 2 — Entrada manual

- novo ordenamento;
- adicionar objeto digitando;
- resolver rua no banco;
- agrupar objetos;
- listar objetos/ruas.

## Fase 3 — Coordenadas

- utilizar geometria/coordenada existente;
- fallback de geocodificação;
- cache/persistência;
- validação em Campos dos Goytacazes.

## Fase 4 — Algoritmo

- ponto inicial do CDD;
- Haversine;
- nearest neighbor;
- ordem estritamente por vizinho mais próximo;
- gerar lista ordenada.

## Fase 5 — Correção manual

- drag-and-drop;
- ordem sugerida;
- ordem final.

## Fase 6 — Voz

- reaproveitar infraestrutura existente;
- integrar ao mesmo pipeline da digitação.

## Fase 7 — Scanner

- câmera;
- leitura da etiqueta;
- resolução;
- fallback para complemento manual/voz.

## Fase 8 — Refinamento

- mobile;
- mensagens;
- erros;
- loading;
- feedback rápido;
- testes finais.

Não é obrigatório seguir exatamente esta ordem se a arquitetura atual indicar uma sequência melhor, mas explique qualquer alteração significativa.

---

# 42. CRITÉRIO DE SUCESSO DO MVP

Considerarei a primeira versão bem-sucedida se um carteiro conseguir:

1. abrir Ordenamento;
2. iniciar novo ordenamento;
3. cadastrar várias encomendas;
4. misturar ruas de diferentes distritos;
5. utilizar digitação inicialmente;
6. sistema identificar e agrupar as ruas;
7. obter coordenadas suficientes;
8. clicar em Gerar ordenamento;
9. receber uma lista lógica partindo do CDD Campos;
10. reorganizar manualmente a lista se quiser.

O objetivo NÃO é substituir um sistema profissional de navegação.

O objetivo é:

**reduzir o tempo de preparação e ordenamento das encomendas.**

Se essa hipótese funcionar bem no uso real, evoluiremos a precisão depois.

---

# 43. COBERTURA DE COORDENADAS — LOTE MANUAL DE 07/09/2026

A planilha operacional continua pendente e não é necessária para corrigir a
cobertura geográfica. Antes do lote manual, a base local tinha 301 ruas ativas
sem geometria utilizável, considerando `NULL` e texto vazio. O arquivo novo
`scripts/geometrias_manuais_novo.json` passou pela validação de formato,
coordenadas, IDs únicos e existência no cadastro.

As 228 geometrias novas foram incorporadas ao arquivo canônico
`scripts/geometrias_manuais.json` (315 entradas no total) e aplicadas em
transação nos dois bancos. O aplicador só preenche registros ativos vazios e
nunca sobrescreve uma geometria existente. O resultado auditado é:

- 2.115 ruas ativas em cada banco;
- 73 ruas ativas ainda sem geometria utilizável;
- 228 ruas atualizadas neste lote e 87 entradas manuais anteriores preservadas;
- 13 ruas que nunca apareceram em uma leva de correspondência OSM;
- hash canônico JSONB das geometrias igual nos bancos após a aplicação:
  `7f0371aa84d767a5ed0438572366cf9d`.

A auditoria cadastral apontou uma divergência independente nas ruas de IDs 721
e 722 (nomes diferentes entre local e produção, com bairro, CEP e distrito
iguais). Nenhum desses registros foi alterado pelo lote de geometrias; a
correção cadastral deve ser tratada separadamente.

O relatório atual está em `scripts/relatorio_ruas_sem_geometria.csv`, com as 73
pendências. O retrato das 301 pendências antes deste lote foi preservado em
`scripts/relatorio_ruas_sem_geometria_antes_lote_2026-09-07.csv`. A lista atual
das 13 ruas sem qualquer correspondência OSM está em
`scripts/ruas_sem_nenhum_match.csv` e a lista anterior, com 17 registros, foi
preservada em `scripts/ruas_sem_nenhum_match_antes_lote_2026-09-07.csv`.

Cada lote futuro deve ser validado por limites geográficos, geometria não
degenerada e comparação do hash entre os bancos antes e depois. CEP ou ponto
aproximado só deve ser aceito quando a relação com a rua for inequívoca e
precisa ficar identificado como coordenada aproximada. Nenhuma rua deve
receber o centro do distrito ou do CDD como substituto.

## Ferramenta manual

O arquivo `scripts/desenhar-ruas-manual.html` agora funciona como editor local
das 73 pendências restantes. Ele incorpora os dados do relatório, permite filtrar por
nome/bairro/distrito/CEP, consultar até cinco resultados do OpenStreetMap,
desenhar múltiplos segmentos, mostrar a coordenada representativa e retomar o
progresso salvo no navegador. O JSON exportado continua compatível com
`scripts/aplicar_geometria_manual.py`.

O aplicador valida limites geográficos e a estrutura `MultiLineString`, ignora
duplicidades e registros inexistentes e só preenche ruas cujo campo ainda está
vazio. Geometrias existentes nunca são sobrescritas automaticamente.

## Revisão de duplicidades por nome — 07/09/2026

Como Campos dos Goytacazes possui ruas com o mesmo nome em bairros diferentes,
o casamento OSM somente por nome-base pode associar trechos de regiões distintas
ao mesmo cadastro. A auditoria local encontrou 97 grupos de nomes exatos
repetidos (232 cadastros) e 127 chaves-base com contextos diferentes. O caso
`NOSSA SENHORA DA PENHA` reúne três cadastros, dois bairros e três CEPs com a
mesma geometria gravada.

Foi criado o relatório somente leitura
`scripts/auditar_duplicidades_geometrias.py`. Ele registra bairro, distrito,
CEP, hash da geometria, extensão do traçado e distância entre os centros para
priorizar a revisão. Os resultados ficam em
`scripts/relatorio_duplicidades_geometrias.csv`,
`scripts/relatorio_duplicidades_geometrias_grupos.csv` e
`scripts/relatorio_ambiguidades_casamento_osm.csv`.

O importador OSM agora preserva cada way e separa os segmentos em componentes
geograficamente contínuos. Correspondências com confiança alta só são gravadas
quando há um componente único; nomes com vários componentes ficam em
`scripts/revisao_matches_ambiguos.csv`. O aplicador de revisões exige o índice
`componente` para uma escolha ambígua. Nenhuma geometria do banco foi alterada
por esta auditoria.

### Revisor visual de duplicidades — 07/09/2026

Foi criada a ferramenta local `scripts/revisar-duplicidades.html`. Ela carrega
`scripts/relatorio_duplicidades_geometrias.csv`, agrupa os cadastros repetidos,
consulta somente o nome selecionado no Overpass e separa os ways em
componentes contínuos no navegador. O operador escolhe o componente que
corresponde ao bairro/CEP do cadastro e exporta
`decisoes_duplicidades_osm.json`.

Quando a revisão estiver concluída, o aplicador recebe esse arquivo com
`python scripts/aplicar_revisao_osm.py --arquivo decisoes_duplicidades_osm.json`.
As credenciais são lidas do ambiente ou de `Backend/.env`.

O revisor não grava no banco. O JSON exportado é compatível com
`aplicar_revisao_osm.py`, que exige o campo `componente` para escolhas
ambíguas. A consulta pode depender da disponibilidade do Overpass; falhas de
timeout não alteram o relatório nem os dados locais.

Pendência identificada: um componente OSM pode ter o nome correto, mas cobrir
um trecho maior que o cadastro postal. Esses casos não devem ser aceitos como
um todo; será necessário revisar e recortar a geometria por segmento antes da
aplicação.
