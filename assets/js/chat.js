const WORKER_URL = "https://flat-rice-6724.sauloleonardo1987.workers.dev";
const historico = [];

// Função de busca de CEP com Fallback (BrasilAPI)
async function buscarCEP(cep) {
    const cepLimpo = cep.replace(/\D/g, '');
    if (cepLimpo.length !== 8) return null;

    try {
        // 1ª Tentativa: ViaCEP
        let res = await fetch(`https://viacep.com.br/ws/${cepLimpo}/json/`);
        let data = await res.json();

        if (data.erro) throw new Error("CEP não encontrado");
        return data;
    } catch (error) {
        console.warn("ViaCEP falhou ou não encontrou. Tentando BrasilAPI...");
        try {
            // 2ª Tentativa: BrasilAPI
            const res = await fetch(`https://brasilapi.com.br/api/cep/v1/${cepLimpo}`);
            const data = await res.json();
            
            if (res.ok) {
                return {
                    logradouro: data.street || "",
                    bairro: data.neighborhood || "",
                    localidade: data.city || "",
                    uf: data.state || "",
                    cep: data.cep || ""
                };
            }
            return null;
        } catch {
            return null;
        }
    }
}

// Função de busca por Nome de Rua em Campos dos Goytacazes
async function buscarPorRua(texto) {
    // Limpa prefixos para melhorar a busca no banco de dados
    const buscaLimpa = texto.toLowerCase()
        .replace(/^(rua|avenida|av\.?|travessa|alameda|praça|estrada)\s+/i, '')
        .trim();

    try {
        const res = await fetch(`https://viacep.com.br/ws/RJ/Campos%20dos%20Goytacazes/${encodeURIComponent(buscaLimpa)}/json/`);
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
    return match ? match[0].trim() : null;
}

function toggleChat() {
    const win = document.getElementById('chat-window');
    win.classList.toggle('aberto');
    if (win.classList.contains('aberto')) {
        document.getElementById('chat-badge') && (document.getElementById('chat-badge').style.display = 'none');
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

    // Lógica de Identificação: CEP ou Nome de Rua
    if (cepEncontrado) {
        const dados = await buscarCEP(cepEncontrado);
        if (dados) {
            infoCEP = `\n\n[Dados do CEP ${cepEncontrado}: Logradouro: ${dados.logradouro}, Bairro: ${dados.bairro}, Cidade: ${dados.localidade}, UF: ${dados.uf}]`;
        } else {
            infoCEP = `\n\n[O CEP ${cepEncontrado} não foi localizado nas bases oficiais.]`;
        }
    } else {
        // Tenta pegar o nome da rua ou assume o texto todo se for uma mensagem curta e direta
        let nomeParaBusca = extrairNomeRua(texto) || (texto.length > 3 && texto.length < 60 ? texto : null);

        if (nomeParaBusca) {
            const resultados = await buscarPorRua(nomeParaBusca);
            if (resultados && resultados.length > 0) {
                infoCEP = `\n\n[Endereços encontrados em Campos: ${resultados.slice(0, 5).map(r =>
                    `CEP: ${r.cep} - ${r.logradouro}, ${r.bairro}`
                ).join(' | ')}]`;
            } else {
                infoCEP = `\n\n[Busca por "${nomeParaBusca}" não retornou resultados em Campos dos Goytacazes.]`;
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
Ajude os carteiros com informações sobre CEPs, endereços e logística. 
Os distritos postais da região vão de 601 a 609.
Use as informações entre colchetes [] para dar respostas precisas. 
Se a busca falhar, oriente o usuário a verificar se o logradouro é novo.`
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
        adicionarMsg('Ocorreu um erro ao processar sua mensagem. Tente novamente.', 'bot');
        console.error(erro);
    }
}

// ===== Lógica de Arrastar Janela =====
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

// ===== Eventos de Botão e Voz =====
document.getElementById('chat-btn').addEventListener('click', toggleChat);

const btnMic = document.getElementById('chat-mic');
const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;

if (SpeechRecognition) {
    const recognition = new SpeechRecognition();
    recognition.lang = 'pt-BR';
    
    recognition.onresult = (e) => {
        const texto = e.results[0][0].transcript;
        document.getElementById('chat-input').value = texto;
        enviarMensagem();
    };

    btnMic.addEventListener('click', () => {
        recognition.start();
        btnMic.style.background = '#ff4d4d';
        recognition.onend = () => btnMic.style.background = '';
    });
} else {
    btnMic.style.display = 'none';
}