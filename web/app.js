async function refreshApplicationState() {
    try {
        const response = await fetch('/api/state');
        const state = await response.json();

        document.querySelector('#count').textContent = `${state.killed}/${state.total}`;
        document.querySelector('#death-count').textContent = state.deaths;
        const deathInput = document.querySelector('#death-input');
        if (document.activeElement !== deathInput) {
            deathInput.value = state.deaths;
        }
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

async function setDeaths(deaths) {
    if (!Number.isInteger(deaths) || deaths < 0) {
        return;
    }

    await fetch('/api/deaths', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deaths })
    });
    await refreshApplicationState();
}

document.querySelector('#death-plus').onclick = () => {
    const currentDeaths = Number(document.querySelector('#death-count').textContent);
    setDeaths(currentDeaths + 1);
};

document.querySelector('#death-minus').onclick = () => {
    const currentDeaths = Number(document.querySelector('#death-count').textContent);
    setDeaths(Math.max(0, currentDeaths - 1));
};

document.querySelector('#death-save').onclick = () => {
    const updatedDeaths = Number(document.querySelector('#death-input').value);
    setDeaths(updatedDeaths);
};

document.querySelector('#death-reset').onclick = () => setDeaths(0);
document.querySelectorAll('.tab').forEach(button => {
    button.onclick = () => {
        document.querySelectorAll('.tab').forEach(tab => {
            tab.classList.toggle('active', tab === button);
        });

        document.querySelectorAll('.panel').forEach(panel => {
            panel.hidden = panel.id !== button.dataset.tab;
        });
    };
});

document.querySelectorAll('button').forEach(button => {
    button.addEventListener('click', () => {
        button.classList.add('is-pressed');
        window.setTimeout(() => button.classList.remove('is-pressed'), 250);
    });
});

setInterval(refreshApplicationState, 1000);
refreshApplicationState();
