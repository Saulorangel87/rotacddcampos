export const APP_VERSION = '1.6.0'

export const UPDATE_NOTES = [
  {
    title: 'Busca por voz mais tolerante',
    description: 'Pequenos erros de transcrição, como Niwton/Newton e Vitor/Victor, agora exibem sugestões do cadastro para confirmação, inclusive no Zé Rota.',
  },
  {
    title: 'Scanner focado no CEP',
    description: 'A câmera ignora códigos de rastreio e só aceita o código de barras numérico com oito dígitos do CEP, mantendo a leitura aberta até encontrar um código válido.',
  },
  {
    title: 'Pesquisa por voz no cabeçalho',
    description: 'O campo Buscar rua no mapa ganhou microfone para iniciar a pesquisa sem digitação.',
  },
  {
    title: 'Acesso por perfil e login',
    description: 'Administradores e colaboradores veem apenas as funções permitidas para o seu perfil, e todo novo login começa no Mapa Geral.',
  },
  {
    title: 'Novidades após o login',
    description: 'Esta tela aparece uma vez por usuário em cada versão, logo depois da entrada no sistema, para apresentar as melhorias da entrega.',
  },
  {
    title: 'Identidade visual do CDD',
    description: 'O cabeçalho, a navegação e o ponto de partida foram refinados para identificar a unidade Campos dos Goytacazes, com ícones próprios para cada perfil.',
  },
  {
    title: 'Ações com melhor contraste',
    description: 'Botões de geração, recálculo e ajuste mantêm texto legível também durante o foco e o movimento do mouse.',
  },
  {
    title: 'Busca de ruas refinada',
    description: 'A pesquisa entende acentos, tipos de logradouro, abreviações, prefixos, variações de digitação e CEP com ou sem hífen, priorizando o cadastro mais exato e tratando ambiguidades para conferência.',
  },
  {
    title: 'Leitura por voz, câmera e CEP',
    description: 'Lance encomendas por voz ou pelo código de barras da etiqueta; quando a leitura traz um CEP, ele é convertido na rua correspondente do cadastro da unidade.',
  },
  {
    title: 'Scanner mais seguro',
    description: 'O painel de câmera inicia e encerra corretamente, com cancelamento limpo e mensagens de feedback para cada leitura.',
  },
  {
    title: 'Confirmação mais clara',
    description: 'Resultados parciais ou ambíguos mostram as opções do cadastro para que a rua correta seja confirmada antes de entrar na rota.',
  },
  {
    title: 'Salvamento mais confiável',
    description: 'A comunicação com a API foi ajustada para confirmar ruas e guardar correções do ordenamento sem falhas de comunicação.',
  },
  {
    title: 'Ordenamento passo a passo',
    description: 'A sequência parte sempre do CDD e escolhe, a cada etapa, a rua restante mais próxima usando os pontos reais do traçado quando disponíveis.',
  },
  {
    title: 'Proximidade pelo traçado real',
    description: 'Quando uma rua possui vários pontos de geometria, o cálculo usa o ponto mais próximo da posição atual e atualiza a posição para essa transição, evitando que uma média distante distorça a rota.',
  },
  {
    title: 'Resultado determinístico',
    description: 'Empates de distância recebem um critério fixo do cadastro, e a sequência não é reordenada por uma otimização global que possa afastar ruas vizinhas.',
  },
  {
    title: 'Correções que permanecem',
    description: 'Ajustes manuais podem ser salvos como sequência habitual do próprio usuário e são preservados ao gerar novamente a ordem.',
  },
  {
    title: 'Nova carga com horário correto',
    description: 'Limpar a lista inicia uma nova carga e atualiza a data e a hora exibidas no ordenamento.',
  },
  {
    title: 'Instalação no celular',
    description: 'A opção de instalar aparece somente em dispositivos móveis e fica oculta quando o site já está instalado.',
  },
]
