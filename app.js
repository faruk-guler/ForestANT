let currentData = null;

const contentContainer = document.getElementById('monitor-content');
const lastUpdatedEl = document.getElementById('last-updated');

async function fetchData() {
    try {
        // Cache busting with timestamp
        const response = await fetch(`data.json?t=${Date.now()}`);
        if (!response.ok) throw new Error('Could not load data template');
        currentData = await response.json();
        render();
    } catch (error) {
        console.error('Error:', error);
        contentContainer.innerHTML = `<div class="error">Error loading data: ${error.message}</div>`;
    }
}

function render() {
    if (!currentData || !currentData.items) {
        contentContainer.innerHTML = '<div class="loader">Waiting for data...</div>';
        return;
    }

    if (currentData.items.length === 0) {
        contentContainer.innerHTML = '<div class="loader">No domains currently configured. Please add domains to domains.txt</div>';
        return;
    }

    // Using en-GB for consistent date format
    lastUpdatedEl.textContent = currentData.last_updated ? new Date(currentData.last_updated).toLocaleString('en-GB') : '--';
    renderTable(currentData.items);
}

function renderTable(items) {
    contentContainer.innerHTML = `
        <div class="table-container">
            <table>
                <thead>
                    <tr>
                        <th>Domain Name</th>
                        <th>Domain Expiry</th>
                        <th>Domain Left</th>
                        <th>SSL Expiry</th>
                        <th>SSL Left</th>
                    </tr>
                </thead>
                <tbody>
                    ${items.map(item => {
                        const dColor = getColorClass(item.domain.days_left);
                        const sColor = getColorClass(item.ssl.days_left);
                        
                        const domainText = item.domain.days_left === 'Error' ? 'Error' : (item.domain.days_left >= 0 ? item.domain.days_left + ' Days' : 'Expired');
                        const sslText = item.ssl.days_left === 'Error' ? 'Error' : (item.ssl.days_left >= 0 ? item.ssl.days_left + ' Days' : 'Expired');
                        
                        return `
                            <tr>
                                <td><strong>${item.domain_name}</strong></td>
                                <td>${item.domain.expiry ? new Date(item.domain.expiry).toLocaleDateString('en-GB') : '-'}</td>
                                <td class="${dColor}"><span class="status-pill"><span class="pulse-dot"></span>${domainText}</span></td>
                                <td>${item.ssl.expiry ? new Date(item.ssl.expiry).toLocaleDateString('en-GB') : '-'}</td>
                                <td class="${sColor}"><span class="status-pill"><span class="pulse-dot"></span>${sslText}</span></td>
                            </tr>
                        `;
                    }).join('')}
                </tbody>
            </table>
        </div>
    `;
}

function getColorClass(days) {
    if (days === 'Error' || days < 0) return 'color-critical';
    if (days < 15) return 'color-critical';
    if (days < 30) return 'color-warning';
    return 'color-ok';
}

// Initial Load
fetchData();
setInterval(fetchData, 300000); // 5 mins
