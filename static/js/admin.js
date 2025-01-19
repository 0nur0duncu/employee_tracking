document.addEventListener('DOMContentLoaded', function() {
    let allWorks = [];
    let allEmployees = [];
    let currentFilter = 'all';
    let selectedDate = new Date();
    selectedDate.setHours(0, 0, 0, 0);

    // Initialize Flatpickr for date picker with Turkish locale
    flatpickr.localize(flatpickr.l10ns.tr);
    const datePicker = flatpickr("#datePicker", {
        dateFormat: "d F Y",
        defaultDate: selectedDate,
        locale: {
            ...flatpickr.l10ns.tr,
            months: {
                longhand: ['Ocak', 'Şubat', 'Mart', 'Nisan', 'Mayıs', 'Haziran', 'Temmuz', 'Ağustos', 'Eylül', 'Ekim', 'Kasım', 'Aralık']
            }
        },
        onChange: function(selectedDates) {
            selectedDate = selectedDates[0];
            loadWorks();
        }
    });

    // Date navigation buttons
    document.getElementById('prevDay').addEventListener('click', () => {
        selectedDate.setDate(selectedDate.getDate() - 1);
        datePicker.setDate(selectedDate);
        loadWorks();
    });

    document.getElementById('nextDay').addEventListener('click', () => {
        selectedDate.setDate(selectedDate.getDate() + 1);
        datePicker.setDate(selectedDate);
        loadWorks();
    });

    document.getElementById('today').addEventListener('click', () => {
        selectedDate = new Date();
        selectedDate.setHours(0, 0, 0, 0);
        datePicker.setDate(selectedDate);
        loadWorks();
    });

    // Load employees and their stats
    async function loadEmployees() {
        try {
            const response = await fetch('/api/employees');
            allEmployees = await response.json();
            
            const tbody = document.getElementById('employeesTableBody');
            tbody.innerHTML = '';

            for (const employee of allEmployees) {
                const statsResponse = await fetch(`/api/work-stats/${employee.id}`);
                const stats = await statsResponse.json();

                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${employee.name}</td>
                    <td>${stats.averageVideoDuration || '-'}</td>
                    <td>${stats.averageSoftwareDuration || '-'}</td>
                    <td>${stats.totalWorks}</td>
                    <td>
                        <button class="btn btn-danger btn-sm" onclick="deleteEmployee('${employee.id}')">Sil</button>
                    </td>
                `;
                tbody.appendChild(tr);
            }

            // After loading employees, refresh the works display
            await loadWorks();
        } catch (error) {
            console.error('Error:', error);
        }
    }

    // Handle employee creation
    document.getElementById('saveEmployee').addEventListener('click', async function() {
        const name = document.getElementById('employeeName').value;
        if (!name) {
            alert('Lütfen personel adını girin!');
            return;
        }

        try {
            const response = await fetch('/api/employees', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ name })
            });

            if (!response.ok) {
                throw new Error('Network response was not ok');
            }

            const modal = bootstrap.Modal.getInstance(document.getElementById('addEmployeeModal'));
            modal.hide();
            document.getElementById('employeeName').value = '';
            await loadEmployees(); // This will also refresh works
            alert('Personel başarıyla eklendi!');
        } catch (error) {
            console.error('Error:', error);
            alert('Personel eklenirken bir hata oluştu!');
        }
    });

    // Delete employee function
    window.deleteEmployee = async function(employeeId) {
        if (!confirm('Bu personeli silmek istediğinize emin misiniz?')) {
            return;
        }

        try {
            const response = await fetch(`/api/employees/${employeeId}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                throw new Error('Network response was not ok');
            }

            await loadEmployees(); // This will also refresh works
            alert('Personel başarıyla silindi!');
        } catch (error) {
            console.error('Error:', error);
            alert('Personel silinirken bir hata oluştu!');
        }
    };

    // Load works and timeline
    async function loadWorks() {
        try {
            const response = await fetch('/api/works');
            const works = await response.json();
            
            // Filter works for selected date
            allWorks = works.filter(work => {
                const workDate = new Date(work.startTime);
                workDate.setHours(0, 0, 0, 0);
                return workDate.getTime() === selectedDate.getTime();
            });

            updateStatistics();
            updateTimeTable();
            filterWorks(currentFilter);
        } catch (error) {
            console.error('Error:', error);
        }
    }

    // Update time table
    function updateTimeTable() {
        const tbody = document.getElementById('timeTableBody');
        tbody.innerHTML = '';

        // Create a row for each employee
        allEmployees.forEach(employee => {
            const tr = document.createElement('tr');
            
            // Add employee name cell
            const nameCell = document.createElement('td');
            nameCell.textContent = employee.name;
            tr.appendChild(nameCell);

            // Add cells for each hour
            for (let hour = 9; hour <= 18; hour++) {
                const td = document.createElement('td');
                
                // Find works for this employee in this hour
                const hourWorks = allWorks.filter(work => {
                    const startTime = new Date(work.startTime);
                    const endTime = work.endTime ? new Date(work.endTime) : new Date();
                    const workStartHour = startTime.getHours();
                    const workEndHour = endTime.getHours();
                    
                    return work.employeeId === employee.id && 
                           ((workStartHour === hour) || (workEndHour === hour) || 
                            (workStartHour < hour && workEndHour > hour));
                });

                // Add work bars
                hourWorks.forEach(work => {
                    const startTime = new Date(work.startTime);
                    const endTime = work.endTime ? new Date(work.endTime) : new Date();
                    
                    // Calculate position and width within the cell
                    let leftPercent = 0;
                    let widthPercent = 100;
                    
                    if (startTime.getHours() === hour) {
                        leftPercent = (startTime.getMinutes() / 60) * 100;
                        widthPercent = 100 - leftPercent;
                    }
                    
                    if (endTime.getHours() === hour) {
                        widthPercent = (endTime.getMinutes() / 60) * 100;
                    }
                    
                    if (startTime.getHours() === hour && endTime.getHours() === hour) {
                        leftPercent = (startTime.getMinutes() / 60) * 100;
                        widthPercent = ((endTime.getMinutes() - startTime.getMinutes()) / 60) * 100;
                    }

                    const bar = document.createElement('div');
                    bar.className = `work-bar ${work.workType} ${work.status}`;
                    bar.style.left = leftPercent + '%';
                    bar.style.width = widthPercent + '%';
                    
                    // Create detailed tooltip
                    const startTimeStr = startTime.toLocaleTimeString('tr-TR', {hour: '2-digit', minute: '2-digit'});
                    const endTimeStr = work.endTime ? 
                        endTime.toLocaleTimeString('tr-TR', {hour: '2-digit', minute: '2-digit'}) : 
                        'Devam Ediyor';
                    
                    const duration = work.endTime ? 
                        formatDuration((endTime - startTime) / (1000 * 60)) : 
                        'Devam Ediyor';
                    
                    bar.title = `${work.workType === 'software' ? 'Yazılım' : 'Video'}
Başlangıç: ${startTimeStr}
${work.endTime ? 'Bitiş: ' + endTimeStr : 'Devam Ediyor'}
${work.endTime ? 'Süre: ' + duration : ''}`;

                    td.appendChild(bar);
                });

                tr.appendChild(td);
            }

            tbody.appendChild(tr);
        });
    }

    // Update statistics
    function updateStatistics() {
        const totalWorks = allWorks.length;
        const completedWorks = allWorks.filter(work => work.status === 'completed').length;
        const activeWorks = allWorks.filter(work => work.status === 'in_progress').length;

        // Calculate average work completion time
        const completedWorksWithDuration = allWorks.filter(work => 
            work.status === 'completed' &&
            work.duration
        );

        let avgWorkTime = '-';
        if (completedWorksWithDuration.length > 0) {
            const totalMinutes = completedWorksWithDuration.reduce((acc, work) => {
                const duration = parseDuration(work.duration);
                return acc + duration;
            }, 0);
            avgWorkTime = formatDuration(totalMinutes / completedWorksWithDuration.length);
        }

        document.getElementById('totalWorks').textContent = totalWorks;
        document.getElementById('completedWorks').textContent = completedWorks;
        document.getElementById('activeWorks').textContent = activeWorks;
        document.getElementById('avgWorkTime').textContent = avgWorkTime;
    }

    // Filter works
    function filterWorks(filter) {
        currentFilter = filter;
        let filteredWorks = allWorks;

        if (filter === 'active') {
            filteredWorks = allWorks.filter(work => work.status === 'in_progress');
        } else if (filter === 'completed') {
            filteredWorks = allWorks.filter(work => work.status === 'completed');
        }

        displayWorks(filteredWorks);
        
        // Update active state of filter buttons
        document.querySelectorAll('.btn-group .btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.getElementById(`filter${filter.charAt(0).toUpperCase() + filter.slice(1)}`).classList.add('active');
    }

    // Display works in table
    function displayWorks(works) {
        const tbody = document.getElementById('worksTableBody');
        tbody.innerHTML = '';

        if (works.length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td colspan="7" class="text-center text-muted">
                    ${selectedDate.toLocaleDateString('tr-TR', { day: 'numeric', month: 'long', year: 'numeric' })} 
                    tarihine ait kayıt bulunamadı
                </td>
            `;
            tbody.appendChild(tr);
            return;
        }

        works.forEach(work => {
            const tr = document.createElement('tr');
            const startTime = new Date(work.startTime);
            const endTime = work.endTime ? new Date(work.endTime) : null;

            tr.innerHTML = `
                <td>${work.employeeName}</td>
                <td>${work.workType === 'software' ? 'Yazılım' : 'Video'}</td>
                <td>${work.videoLink ? `<a href="${work.videoLink}" target="_blank">Link</a>` : '-'}</td>
                <td>${startTime.toLocaleTimeString('tr-TR')}</td>
                <td>${endTime ? endTime.toLocaleTimeString('tr-TR') : '-'}</td>
                <td>${work.duration || '-'}</td>
                <td>
                    <span class="badge ${work.status === 'completed' ? 'badge-completed' : 'badge-in-progress'}">
                        ${work.status === 'completed' ? 'Tamamlandı' : 'Devam Ediyor'}
                    </span>
                </td>
            `;
            tbody.appendChild(tr);
        });
    }

    // Helper function to parse duration string to minutes
    function parseDuration(duration) {
        const parts = duration.split(':');
        if (parts.length === 3) {
            // Format: HH:MM:SS
            return parseInt(parts[0]) * 60 + parseInt(parts[1]);
        } else if (parts.length === 2) {
            // Format: MM:SS
            return parseInt(parts[0]);
        }
        return 0;
    }

    // Helper function to format minutes to duration string
    function formatDuration(minutes) {
        const hours = Math.floor(minutes / 60);
        const mins = Math.floor(minutes % 60);
        return hours > 0 ? `${hours}s ${mins}dk` : `${mins}dk`;
    }

    // Event listeners for filter buttons
    document.getElementById('filterAll').addEventListener('click', () => filterWorks('all'));
    document.getElementById('filterActive').addEventListener('click', () => filterWorks('active'));
    document.getElementById('filterCompleted').addEventListener('click', () => filterWorks('completed'));

    // Initial load
    loadEmployees(); // This will also load works

    // Refresh data every 30 seconds
    setInterval(() => {
        loadEmployees(); // This will also refresh works
    }, 30000);
}); 