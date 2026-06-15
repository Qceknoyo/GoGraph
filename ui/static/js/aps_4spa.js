document.addEventListener('DOMContentLoaded', () => {
    loadHistory();

   
            
    const btn = document.getElementById('toggle-btn');
    const sidebar = document.getElementById('sidebar');

    btn.addEventListener('click', () => {
        if (window.innerWidth > 768) {
            // на комп-е — просто сворачиваем в узкую полоску
            sidebar.classList.toggle('collapsed');
        } else {
            // на мобилк-е — раскрываем список вниз
            sidebar.classList.toggle('mobile-open');
        }
    });

    // ввод данных с формы для базы данных (бд)
    
    // const dialog = document.getElementById('dbDialog');
    // const form = document.getElementById('dbForm');
    
    
    // if (form) {
    //         form.addEventListener('submit', (e) => {
    //         e.preventDefault();

    //         const data = new FormData(e.target);

            
    //         // обертываем data в URLSearchParams (обычный формадата сложный для него очень)
    //         fetch("/connectDB", {
    //             method: 'POST',
    //             body: new URLSearchParams(data) 
    //         })
    //         .then(response => response.json())
    //         .then(data => {
    //             console.log('Успех:', data);
    //             const pre = document.querySelector('.result-section pre');
    //             if (pre) {
    //                 pre.innerText = JSON.stringify(data, null, 2);
    //             } else {
    //                 console.log('pre не найден, но data есть:', data);
    //             }
    //             alert('Данные для подключения загружены!');
    //         })
    //         .catch(error => {
    //             console.error('Ошибка:', error);
    //             alert('Произошла ошибка при отправке!');
    //         });

    //         if (dialog) {
    //             dialog.close();
    //             console.log('6. Диалог закрыт');
    //         }
    //     });

    // }

    //ввод данных через файл
    const fileElem = document.getElementById('fileElem');
    fileElem.addEventListener("change", async (e) => {
        e.preventDefault();
        
        const files = e.target.files;
        
        if (files.length === 0) return;
        
        console.log(`Выбрано файлов: ${files.length}`);
        
        const formData = new FormData();
        
        //добавить все файлы
        Array.from(files).forEach(file => {
            formData.append('files', file);
        });
        
        // сразу один запрос на все файлы
        fetch('/upload', {  
            method: 'POST',
            body: formData
        })
        .then(response => response.json())  // ожидание jsonчика
        .then(data => {
            data4Dashboard(data);
            loadHistory();
            console.log('Успех:', data);
            const pre = document.querySelector('.result-section pre');

            if (pre) {
                pre.innerText = JSON.stringify(data, null, 2);
            }

            // АНАЛИТИКА 
            const analysis = data.analysis;

            if (analysis) {

                alert(`
        Исследование завершено

        Макс. напряжение: ${analysis.max_stress.toFixed(2)} МПа
        Среднее напряжение: ${analysis.avg_stress.toFixed(2)} МПа
        Точка разрушения: ${analysis.break_point.toFixed(2)} %
        Точек измерений: ${analysis.total_points}
                `);
            }
        })
            
            
        .catch(error => {
            console.error('Ошибка:', error);
            alert('Ошибка загрузки');
        });
    });


//функция чтобы нормально даты отображались 
function formatDate(dateString) {
    return new Date(dateString).toLocaleString('ru-RU',{ timeZone: 'Europe/Moscow' });
}



// функция загрузки истории
async function loadHistory() {

    try {

        const response = await fetch('/history');

        const result = await response.json();

        console.log(result);

        const container = document.getElementById('historyContainer');

        container.innerHTML = '';

        if (!result.data || result.data.length === 0) {

            container.innerHTML =
                '<p class="empty-message">📭 История исследований пуста</p>';

            return;
        }

        let cardsHtml = '';

        result.data.forEach(exp => {

            cardsHtml += `

            <div
                class="history-card"
                data-date="${exp.created_at}"
            >

                <h3>
                    🧪 Исследование #${escapeHtml(exp.id)}
                </h3>

                <p>
                    <strong>Макс. напряжение:</strong>
                    ${Number(exp.max_stress).toFixed(2)} МПа
                </p>

                <p>
                    <strong>Среднее напряжение:</strong>
                    ${Number(exp.avg_stress).toFixed(2)} МПа
                </p>

                <p>
                    <strong>Количество точек:</strong>
                    ${escapeHtml(exp.total_points)}
                </p>

                <p>
                    <strong>Дата:</strong>
                    ${formatDate(exp.created_at)}
                </p>

            </div>
        `;
        });

        container.innerHTML = cardsHtml;

    } catch (err) {

        console.error(err);

        document.getElementById('historyContainer').innerHTML =
            '<p class="error-message">⚠️ Ошибка загрузки истории</p>';
    }
}
window.loadHistory = loadHistory;

// защита от xss
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// все кнопки меню тут
const navButtons = document.querySelectorAll(".nav-link");        
// ф-ия скрытие страницы
function switchPage(pageName) {
    // скрытие страниц
    const allPages = document.querySelectorAll('.page');
    allPages.forEach(page => {
        page.classList.remove('active');
    });
            
    // показ нужной страницы
    const activePage = document.getElementById(`page-${pageName}`);
    if (activePage) {
        activePage.classList.add('active');
    }
            
    //обновление активной кнопки в меню
    navButtons.forEach(btn => {
        btn.classList.remove('active');
        if (btn.getAttribute('data-page') === pageName) {
            btn.classList.add('active');
        }
    });
}
        
        // обработчик
navButtons.forEach(button => {
    button.addEventListener('click', function(event) {
        event.preventDefault(); 
        const pageName = this.getAttribute('data-page');
        switchPage(pageName);
    });
});
      



//передача сессии по куке
async function showSyncLink() {
    const response = await fetch('/sync/generate');
    const data = await response.json();
    const link = `${window.location.origin}/sync?token=${data.token}`;
    await navigator.clipboard.writeText(link);
    alert('Ссылка скопирована (действует 15 минут):\n\n' + link);
}

//просто обработчик чтобы сменить стандартное поведение бразуера
const SyncLinkElem = document.getElementById('SyncLink');
SyncLinkElem.addEventListener("click", async (e) => {
    e.preventDefault();
    showSyncLink();
})


//фильтрация сугубо на фронте
document.getElementById('historyDateFilter').addEventListener('change', function() {

    const selected = this.value;

    document
    .querySelectorAll('.history-card')
    .forEach(card => {

        const cardDate =
            card.dataset.date;

        if (!selected || cardDate.startsWith(selected)) {

            card.style.display = 'block';

        } else {

            card.style.display = 'none';
        }
    });
});

// загружаем первую страницу по умолчанию 
// находим активную кнопку и показываем её страницу
const activeButton = document.querySelector('.nav-link.active');
if (activeButton) {
    const defaultPage = activeButton.getAttribute('data-page');
    if (defaultPage) {
        switchPage(defaultPage);
    } else {
        switchPage('monitoring'); // страница по умолчанию
    }
    } else {
        switchPage('monitoring');
    }

// инициализация графика через чартджс
const ctx = document.getElementById('liveChart').getContext('2d');
const liveChartID = new Chart(ctx, {
    type: 'line',
    data: {
        labels: [],  //x-axis labels
        datasets: [{ //massive of lines on graphic
            label: 'Stress (МПа)', //line name
            data: [], // y-axis value
            borderColor: '#00d2ff',
            backgroundColor: 'rgba(0, 210, 255, 0.05)',
            borderWidth: 2,
            tension: 0.4,
            fill: true
        }]
    },
    options: {
        responsive: true, //to fit the window
        maintainAspectRatio: true, //чтобы график не растянулся и не убил оперативку (второй раз)
        scales: {
            x: { 
                title: {
                    display: true,
                    text: 'Strain (Деформация, %)',
                    color: '#a0aec0'
                },
                grid: { display: false },
                ticks: { color: '#a0aec0' } 
            },
            y: { 
                title: {
                    display: true,
                    text: 'Stress (Напряжение, МПа)',
                    color: '#a0aec0'
                },
                beginAtZero: true,
                grid: { color: 'rgba(255, 255, 255, 0.05)' },
                ticks: { color: '#a0aec0' } //color of marks
            }
        },
        plugins: {
            legend: {   
                display: true,
                labels: { color: '#fff' } 
            }
        }
    }
});

let step_timeList = []; 
let strainList = [];
let stressList = [];

function data4Dashboard(measurements) {
    const dstatus = measurements.status;
    const dpayload = measurements.payload;
    for (const fileName in dpayload){
        const measurements = dpayload[fileName];
        measurements.forEach(item => {
            console.log(item.step_time);
            step_timeList.push(item.step_time)
            console.log(item.strain); 
            strainList.push(item.strain)
            console.log(item.stress);
            stressList.push(item.stress)
        });
    }
    
}

let streamInterval;
let currentIndex = 0;

function simulateStream() {
    let strain = 0;
    let stress = 0;
    currentIndex = 0; // сбрасывание индекса при каждом запуске
    
    clearInterval(streamInterval);

    
    if (strainList.length === 0) {
        alert("Данных нет, провожу тестовое испытание");
    }

    streamInterval = setInterval(() => {
        if (strainList.length === 0) {
            // отрисовка если ничего не скормили
            strain += 0.5;
            stress += (strain < 5) ? (Math.random() * 2 + 5) : (Math.random() * 1 - 0.2);

            liveChartID.data.labels.push(strain.toFixed(1) + "%");
            liveChartID.data.datasets[0].data.push(stress.toFixed(2));

            if (liveChartID.data.labels.length > 20) {
                liveChartID.data.labels.shift();
                liveChartID.data.datasets[0].data.shift();
            }
            liveChartID.update('none');

            if (strain >= 15) {
                clearInterval(streamInterval);
                alert("Материал разрушен! Испытание завершено.");
                document.querySelector('.btn-start').innerText = "🛑 Разрыв";
            }
        } else {
            // отрисовка готового
            if (currentIndex < strainList.length) {
                // добавление одного элемента за раз
                let value = strainList[currentIndex];
                
                liveChartID.data.labels.push(value.toFixed(1) + "%");
                liveChartID.data.datasets[0].data.push(stressList[currentIndex].toFixed(2));

                if (liveChartID.data.labels.length > 20) {
                    liveChartID.data.labels.shift();
                    liveChartID.data.datasets[0].data.shift();
                }

                liveChartID.update('none');
                currentIndex++; // увеличение индекса для следующего тика через 500мс
            } else {
                // данные в массиве закончились
                clearInterval(streamInterval);
                alert("Испытание завершено.");
                document.querySelector('.btn-start').innerText = "🛑 Разрыв";
            }
        }
    }, 500);
}

// button start
document.querySelector('.btn-start').addEventListener('click', function() {
    this.innerText = "Запись...";
    this.style.background = "#e53e3e";
    simulateStream();
});

// button clear
document.querySelector('.btn-clear').addEventListener('click', function(e) {
    e.preventDefault();
    clearInterval(streamInterval);
    const startBtn = document.querySelector('.btn-start');
    startBtn.innerText = "🟢 Запустить замер";
    startBtn.style.background = ""; 
    liveChartID.data.labels = [];
    liveChartID.data.datasets[0].data = [];
    liveChartID.update();
});



// const openBtn = document.getElementById('openBtn');
// const closeBtn = document.getElementById('closeBtn');

// openBtn.onclick = () => dialog.showModal();

// closeBtn.onclick = () => dialog.close();

// // закрыть при клике на темную область
// dialog.addEventListener('click', (e) => {
//   if (e.target === dialog) dialog.close();
// });

// сама обработка формы

});

// удаление архива
async function clearHistory() {

    if (!confirm('Удалить историю?')) {
        return;
    }

    try {

        const response = await fetch('/clearHistory', {
            method: 'POST'
        });

        const result = await response.json();

        if (result.status === 'ok') {

            alert('История удалена');

            loadHistory();
        }

    } catch (err) {

        console.error(err);

        alert('Ошибка удаления');
    }
}

window.clearHistory = clearHistory; 

//функция экспорта
function exportReport() {

    window.location.href = '/export';
}