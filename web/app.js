async function refreshApplicationState() {
    try {
        const response = await fetch('/api/state');
        const state = await response.json();

        document.querySelector('#count').textContent = `${state.killed}/${state.total}`;
        document.querySelector('#status').textContent = state.connected
            ? 'Игра подключена'
            : 'Ожидание подключения к игре';
        const labelInput = document.querySelector('#label');
        if (document.activeElement !== labelInput) {
            labelInput.value = state.label;
        }
    } catch {
        document.querySelector('#status').textContent = 'Сервер не отвечает';
    }
}

async function saveLabel() {
    await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            label: document.querySelector('#label').value
        })
    });

    await refreshApplicationState();
}

document.querySelector('#save-label').onclick = saveLabel;

setInterval(refreshApplicationState, 1000);
refreshApplicationState();
