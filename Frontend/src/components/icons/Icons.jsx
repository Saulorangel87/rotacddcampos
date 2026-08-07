// Ícones de linha (stroke), estilo consistente entre si — mesma abordagem
// de SVG inline já usada no Footer.jsx, evitando adicionar uma lib de ícones
// só pra isso (mantém o bundle pequeno, importa pro PageSpeed).
//
// Todos recebem `size` (px) e repassam o resto das props pro <svg>, então
// dá pra usar className, aria-hidden, etc. normalmente.

const base = {
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.8,
  strokeLinecap: "round",
  strokeLinejoin: "round",
};

function Svg({ size = 20, children, ...props }) {
  return (
    <svg width={size} height={size} {...base} {...props}>
      {children}
    </svg>
  );
}

export function IconeMapa(props) {
  return (
    <Svg {...props}>
      <path d="M9 4 3 6v14l6-2 6 2 6-2V4l-6 2-6-2Z" />
      <path d="M9 4v14M15 6v14" />
    </Svg>
  );
}

export function IconeAjustes(props) {
  return (
    <Svg {...props}>
      <path d="M14.7 6.3 17.7 3.3l3 3-3 3-2.6-.7-.7 2.6L4.3 20.3a1.5 1.5 0 0 1-2.1-2.1L13.3 7l-.6-.7Z" />
      <path d="m14 7 3 3" />
    </Svg>
  );
}

export function IconeRuas(props) {
  return (
    <Svg {...props}>
      <path d="M4 21 10 3h4l6 18" />
      <path d="M8.5 15h7" />
      <path d="M12 3v4M12 9.5v2.5M12 15v6" strokeDasharray="2.2 2.2" />
    </Svg>
  );
}

export function IconeCep(props) {
  return (
    <Svg {...props}>
      <path d="M3 7h18v12H3z" />
      <path d="m3 7 9 7 9-7" />
    </Svg>
  );
}

export function IconeColaboradores(props) {
  return (
    <Svg {...props}>
      <circle cx="9" cy="8" r="3" />
      <path d="M3.5 20a5.5 5.5 0 0 1 11 0" />
      <path d="M16 8.2a3 3 0 1 1 2.2 5.4" />
      <path d="M15.5 14.3c2.6.2 4.8 1.9 5 5.7" />
    </Svg>
  );
}

export function IconeRelatorios(props) {
  return (
    <Svg {...props}>
      <path d="M5 20V10M12 20V4M19 20v-7" />
      <path d="M3 20h18" />
    </Svg>
  );
}

export function IconeUsuarios(props) {
  return (
    <Svg {...props}>
      <circle cx="8" cy="8" r="4.5" />
      <path d="M14.5 12.5 21 19M18.3 15.3l2.7-2.7" />
    </Svg>
  );
}

export function IconeMoto(props) {
  return (
    <Svg {...props}>
      <circle cx="6" cy="17" r="3" />
      <circle cx="18" cy="17" r="3" />
      <path d="M9 17h6l-2-5H8l-1.5 3M13 12l2-3h3M15 9h3.5" />
    </Svg>
  );
}

export function IconeCarro(props) {
  return (
    <Svg {...props}>
      <path d="M4 16V11l2.2-4.5A2 2 0 0 1 8 5.5h8a2 2 0 0 1 1.8 1L20 11v5" />
      <path d="M4 16h16v2a1 1 0 0 1-1 1h-1.5a1 1 0 0 1-1-1v-1h-9v1a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1v-2Z" />
      <circle cx="7.5" cy="16" r="1.4" />
      <circle cx="16.5" cy="16" r="1.4" />
    </Svg>
  );
}

export function IconeBike(props) {
  return (
    <Svg {...props}>
      <circle cx="6" cy="17" r="3" />
      <circle cx="18" cy="17" r="3" />
      <path d="M6 17 10 8h4l3 5M10 8H8M13 17h5" />
    </Svg>
  );
}

export function IconePredio(props) {
  return (
    <Svg {...props}>
      <path d="M5 21V5a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v16" />
      <path d="M13 10h5a1 1 0 0 1 1 1v10" />
      <path d="M8 8h0M8 11h0M8 14h0M8 17h0" strokeWidth="2.6" />
      <path d="M3 21h18" />
    </Svg>
  );
}

export function IconePasta(props) {
  return (
    <Svg {...props}>
      <path d="M3 7a1 1 0 0 1 1-1h4.5l1.5 2H20a1 1 0 0 1 1 1v9a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V7Z" />
    </Svg>
  );
}

export function IconeCaixa(props) {
  return (
    <Svg {...props}>
      <path d="M3 8 12 4l9 4-9 4-9-4Z" />
      <path d="M3 8v9l9 4 9-4V8" />
      <path d="M12 12v9" />
    </Svg>
  );
}

export function IconeEnvelope(props) {
  return (
    <Svg {...props}>
      <path d="M3 6h18v12H3z" />
      <path d="m3 6 9 7 9-7" />
    </Svg>
  );
}

export function IconeBussola(props) {
  return (
    <Svg {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="m14.5 9.5-2 5-5 2 2-5 5-2Z" />
    </Svg>
  );
}

export function IconeFolgas(props) {
  return (
    <Svg {...props}>
      <path d="M4 5h16v15H4z" />
      <path d="M4 9.5h16M8 3v4M16 3v4" />
      <path d="m9 14 2 2 4-4" />
    </Svg>
  );
}

export function IconeEstrela(props) {
  return (
    <Svg {...props}>
      <path d="m12 3 2.6 5.6 6.1.6-4.6 4.1 1.3 6-5.4-3.2L6.6 19l1.3-6-4.6-4.1 6.1-.6L12 3Z" />
    </Svg>
  );
}
