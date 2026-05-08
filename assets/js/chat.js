const WORKER_URL = "https://flat-rice-6724.sauloleonardo1987.workers.dev";
const historico = [];

async function buscarCEP(cep) {
    const cepLimpo = cep.replace(/\D/g, '');
    if (cepLimpo.length !== 8) return null;
    try {
        const res = await fetch(`https://viacep.com.br/ws/${cepLimpo}/json/`);
        const data = await res.json();
        if (data.erro) return null;
        return data;
    } catch {
        return null;
    }
}

async function buscarPorRua(texto) {
    try {
        const res = await fetch(`https://viacep.com.br/ws/RJ/Campos%20dos%20Goytacazes/${encodeURIComponent(texto)}/json/`);
        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0) return null;
        return data;
    } catch {
        return null;
    }
}

function extrairCEP(texto) {
    const match = texto.match(/\d{5}-?\d{3}/);
    return match ? match[0] : null;
}

function extrairNomeRua(texto) {
    const regexRua = /(?:rua|avenida|av\.?|travessa|alameda|praça|estrada)\s+[\wÀ-ÿ\s]+/i;
    const match = texto.match(regexRua);
    if (!match) return null;
    return match[0].trim();
}

function toggleChat() {
    const win = document.getElementById('chat-window');
    win.classList.toggle('aberto');
    if (win.classList.contains('aberto')) {
        document.getElementById('chat-badge').style.display = 'none';
        document.getElementById('chat-input').focus();
    }
}

function adicionarMsg(texto, tipo) {
    const area = document.getElementById('chat-mensagens');
    const div = document.createElement('div');
    div.className = `msg ${tipo}`;
    div.textContent = texto;
    area.appendChild(div);
    area.scrollTop = area.scrollHeight;
    return div;
}

async function enviarMensagem() {
    const input = document.getElementById('chat-input');
    const texto = input.value.trim();
    if (!texto) return;

    input.value = '';
    adicionarMsg(texto, 'user');

    const digitando = adicionarMsg('Digitando...', 'digitando');

    let infoCEP = '';
    const cepEncontrado = extrairCEP(texto);

    if (cepEncontrado) {
        const dados = await buscarCEP(cepEncontrado);
        if (dados) {
            infoCEP = `\n\n[Dados do CEP ${cepEncontrado}: Logradouro: ${dados.logradouro}, Bairro: ${dados.bairro}, Cidade: ${dados.localidade}, UF: ${dados.uf}]`;
        }
    } else {
        const palavrasChave = ['rua', 'avenida', 'av.', 'av ', 'travessa', 'alameda', 'praça', 'estrada'];
        const temRua = palavrasChave.some(p => texto.toLowerCase().includes(p));
        if (temRua) {
            const nomeRua = extrairNomeRua(texto);
            if (nomeRua) {
                const resultados = await buscarPorRua(nomeRua);
                if (resultados && resultados.length > 0) {
                    infoCEP = `\n\n[Endereços encontrados no ViaCEP: ${resultados.slice(0, 5).map(r =>
                        `CEP: ${r.cep} - ${r.logradouro}, ${r.bairro}`
                    ).join(' | ')}]`;
                } else {
                    infoCEP = `\n\n[Busca no ViaCEP não retornou resultados para "${nomeRua}" em Campos dos Goytacazes]`;
                }
            }
        }
    }

    historico.push({ role: 'user', content: texto + infoCEP });

    try {
        const resposta = await fetch(WORKER_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                model: 'llama-3.3-70b-versatile',
                messages: [
                    {
                        role: 'system',
                        content: `Você é um assistente virtual do CDD de Campos dos Goytacazes, RJ. 
Ajude os carteiros com informações sobre CEPs, endereços, bairros e logística da região.
Os distritos postais vão de 601 a 609.
Quando receber dados de um CEP ou endereço entre colchetes [], use essas informações para responder de forma clara.
Se não houver resultados no ViaCEP, oriente o usuário a consultar buscacepinter.correios.com.br.
Responda SEMPRE em português do Brasil, de forma clara e amigável.`
                    },
                    ...historico
                ]
            })
        });

        const data = await resposta.json();
        const mensagemBot = data.choices[0].message.content;

        digitando.remove();
        adicionarMsg(mensagemBot, 'bot');
        historico.push({ role: 'assistant', content: mensagemBot });

    } catch (erro) {
        digitando.remove();
        adicionarMsg('Desculpe, ocorreu um erro. Tente novamente.', 'bot');
        console.error(erro);
    }
}

// ===== CHAT ARRASTÁVEL (cabeçalho) =====
const chatWindow = document.getElementById('chat-window');
const chatHeader = document.getElementById('chat-header');

let isDragging = false;
let offsetX, offsetY;

chatHeader.addEventListener('mousedown', (e) => {
    isDragging = true;
    offsetX = e.clientX - chatWindow.getBoundingClientRect().left;
    offsetY = e.clientY - chatWindow.getBoundingClientRect().top;
    chatHeader.style.cursor = 'grabbing';
});

document.addEventListener('mousemove', (e) => {
    if (!isDragging) return;
    const x = e.clientX - offsetX;
    const y = e.clientY - offsetY;
    const maxX = window.innerWidth - chatWindow.offsetWidth;
    const maxY = window.innerHeight - chatWindow.offsetHeight;
    chatWindow.style.left = `${Math.max(0, Math.min(x, maxX))}px`;
    chatWindow.style.top = `${Math.max(0, Math.min(y, maxY))}px`;
    chatWindow.style.right = 'auto';
    chatWindow.style.bottom = 'auto';
});

document.addEventListener('mouseup', () => {
    isDragging = false;
    chatHeader.style.cursor = 'grab';
});

// ===== BOTÃO CLICÁVEL =====
const chatBtn = document.getElementById('chat-btn');
chatBtn.addEventListener('click', toggleChat);