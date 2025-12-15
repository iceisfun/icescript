// Editor State
let currentScript = null;
let editor = null;

// Monaco Config
require.config({ paths: { 'vs': 'https://unpkg.com/monaco-editor@0.44.0/min/vs' } });

require(['vs/editor/editor.main'], function () {
    // Register generic language for now, or use 'javascript' as placeholder since syntax is similar
    monaco.languages.register({ id: 'icescript' });

    // Define simple tokenizer for coloring (Mock C-like)
    monaco.languages.setMonarchTokensProvider('icescript', {
        tokenizer: {
            root: [
                [/\b(fn|return|if|else|true|false|null|let|while|for)\b/, "keyword"],
                [/[a-z_$][\w$]*/, "identifier"],
                [/"[^"]*"/, "string"],
                [/\d+/, "number"],
                [/\/\/.*/, "comment"],
            ]
        }
    });

    // Define Theme
    monaco.editor.defineTheme('icescript-dark', {
        base: 'vs-dark',
        inherit: true,
        rules: [
            { token: 'keyword', foreground: 'ff7b72' },
            { token: 'string', foreground: 'a5d6ff' },
            { token: 'number', foreground: '79c0ff' },
            { token: 'comment', foreground: '8b949e' },
        ],
        colors: {
            'editor.background': '#0d1117',
        }
    });

    editor = monaco.editor.create(document.getElementById('monaco-editor'), {
        value: '// Select a script or create new one to start coding...',
        language: 'icescript',
        theme: 'icescript-dark',
        automaticLayout: true,
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 14,
        minimap: { enabled: false }
    });

    // Initial Load
    fetchScripts();
});

// API Calls
async function fetchScripts() {
    const res = await fetch('/api/scripts');
    const scripts = await res.json();
    renderScriptList(scripts);
}

async function loadScript(name) {
    const res = await fetch(`/api/scripts/${name}`);
    if (!res.ok) {
        logError(`Failed to load script ${name}`);
        return;
    }
    const content = await res.text();
    editor.setValue(content);
    currentScript = name;
    updateUI();
    log(`Loaded ${name}`);
}

async function saveScript() {
    if (!currentScript) {
        document.getElementById('new-script-dialog').classList.remove('hidden');
        return;
    }
    const content = editor.getValue();
    const res = await fetch(`/api/scripts/${currentScript}`, {
        method: 'POST',
        body: content
    });

    if (res.ok) {
        logSuccess(`Saved ${currentScript}`);
    } else {
        logError(`Failed to save ${currentScript}`);
    }
}

async function deleteScript() {
    if (!currentScript || !confirm(`Delete ${currentScript}?`)) return;

    const res = await fetch(`/api/scripts/${currentScript}`, {
        method: 'DELETE'
    });

    if (res.ok) {
        logSuccess(`Deleted ${currentScript}`);
        currentScript = null;
        editor.setValue('');
        fetchScripts();
        updateUI();
    } else {
        logError(`Failed to delete ${currentScript}`);
    }
}

async function testScript() {
    log('Running test...');
    const content = editor.getValue();
    const res = await fetch('/api/test', {
        method: 'POST',
        body: content
    });

    const result = await res.json();

    // Clear logs mostly? Or keep history. Let's clear for "Test" run.
    document.getElementById('logs').innerHTML = '';

    if (result.Error) {
        if (result.Output) log(result.Output);
        logError(result.Error);
    } else {
        log(result.Output || '(No output)');
        logSuccess('Test completed successfully');
    }
}

// UI Logic
function renderScriptList(scripts) {
    const container = document.getElementById('script-list');
    container.innerHTML = '';
    scripts.forEach(name => {
        const div = document.createElement('div');
        div.className = `script-item ${name === currentScript ? 'active' : ''}`;
        div.textContent = name;
        div.onclick = () => loadScript(name);
        container.appendChild(div);
    });
}

function updateUI() {
    // Update active class in list
    const items = document.querySelectorAll('.script-item');
    items.forEach(el => {
        if (el.textContent === currentScript) el.classList.add('active');
        else el.classList.remove('active');
    });

    document.getElementById('delete-btn').disabled = !currentScript;
}

function log(msg) {
    const div = document.createElement('div');
    div.className = 'log-entry';
    div.textContent = msg;
    document.getElementById('logs').appendChild(div);
    scrollToBottom();
}

function logError(msg) {
    const div = document.createElement('div');
    div.className = 'log-entry log-error';
    div.textContent = msg;
    document.getElementById('logs').appendChild(div);
    scrollToBottom();
}

function logSuccess(msg) {
    const div = document.createElement('div');
    div.className = 'log-entry log-success';
    div.textContent = msg;
    document.getElementById('logs').appendChild(div);
    scrollToBottom();
}

function scrollToBottom() {
    const pane = document.getElementById('output-pane');
    pane.scrollTop = pane.scrollHeight;
}

// Event Listeners
document.getElementById('save-btn').onclick = saveScript;
document.getElementById('delete-btn').onclick = deleteScript;
document.getElementById('test-btn').onclick = testScript;

const newDialog = document.getElementById('new-script-dialog');
document.getElementById('new-btn').onclick = () => newDialog.classList.remove('hidden');
document.getElementById('cancel-new-btn').onclick = () => newDialog.classList.add('hidden');
document.getElementById('confirm-new-btn').onclick = async () => {
    const nameInput = document.getElementById('new-script-name');
    const name = nameInput.value.trim();
    if (!name) return;

    currentScript = name;

    // Only set default content if editor is empty or has default placeholder
    const val = editor.getValue().trim();
    if (!val || val === '// Select a script or create new one to start coding...') {
        editor.setValue('// New script ' + name);
    }

    await saveScript(); // Creates the file

    newDialog.classList.add('hidden');
    nameInput.value = '';
    fetchScripts();
    updateUI();
};
