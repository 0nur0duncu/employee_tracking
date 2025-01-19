document.addEventListener('DOMContentLoaded', function() {
    // Set current date in the header
    const now = new Date();
    document.getElementById('currentDate').textContent = now.toLocaleDateString('tr-TR', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });

    // Load employees
    async function loadEmployees() {
        try {
            const response = await fetch('/api/employees');
            const employees = await response.json();
            
            const select = document.getElementById('employeeSelect');
            select.innerHTML = '<option value="">Seçiniz</option>';
            
            if (employees.length === 0) {
                select.innerHTML = '<option value="">Henüz personel tanımlanmamış</option>';
                select.disabled = true;
                document.getElementById('workForm').querySelectorAll('select, input, button').forEach(el => {
                    el.disabled = true;
                });
                alert('Henüz hiç personel tanımlanmamış. Lütfen yöneticinize başvurun.');
                return;
            }

            select.disabled = false;
            document.getElementById('workForm').querySelectorAll('select, input, button').forEach(el => {
                el.disabled = false;
            });
            
            employees.forEach(employee => {
                const option = document.createElement('option');
                option.value = employee.id;
                option.textContent = employee.name;
                select.appendChild(option);
            });
        } catch (error) {
            console.error('Error:', error);
            alert('Personel listesi yüklenirken bir hata oluştu!');
        }
    }

    // Show/hide video fields based on work type selection
    document.getElementById('workType').addEventListener('change', function(e) {
        const videoFields = document.getElementById('videoFields');
        const videoTypeInputs = document.getElementsByName('videoType');

        if (e.target.value === 'video') {
            videoFields.style.display = 'block';
            videoTypeInputs[0].required = true;
        } else {
            videoFields.style.display = 'none';
            videoTypeInputs[0].required = false;
        }
    });

    // Handle work form submission
    document.getElementById('workForm').addEventListener('submit', async function(e) {
        e.preventDefault();

        const employeeSelect = document.getElementById('employeeSelect');
        const selectedOption = employeeSelect.options[employeeSelect.selectedIndex];

        const formData = {
            employeeId: employeeSelect.value,
            employeeName: selectedOption.text,
            workType: document.getElementById('workType').value,
            startTime: new Date().toISOString(),
            isFirstVideo: document.querySelector('input[name="videoType"]:checked')?.value === 'true' || false
        };

        try {
            const response = await fetch('/api/work', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            if (!response.ok) {
                throw new Error('Network response was not ok');
            }

            alert('İş başarıyla kaydedildi!');
            loadActiveWorks();
            e.target.reset();
        } catch (error) {
            console.error('Error:', error);
            alert('İş kaydedilirken bir hata oluştu!');
        }
    });

    // Load active works for today
    async function loadActiveWorks() {
        try {
            const response = await fetch('/api/works');
            const works = await response.json();
            
            // Filter works for today
            const today = new Date();
            today.setHours(0, 0, 0, 0);
            
            const todayWorks = works.filter(work => {
                const workDate = new Date(work.startTime);
                workDate.setHours(0, 0, 0, 0);
                return workDate.getTime() === today.getTime();
            });

            const activeWorksContainer = document.getElementById('activeWorks');
            activeWorksContainer.innerHTML = '';

            if (todayWorks.length === 0) {
                activeWorksContainer.innerHTML = '<div class="text-center text-muted p-3">Bugün için henüz iş kaydı bulunmuyor.</div>';
                return;
            }

            todayWorks.forEach(work => {
                const workElement = document.createElement('div');
                workElement.className = 'list-group-item list-group-item-action';
                const startTime = new Date(work.startTime);
                const endTime = work.endTime ? new Date(work.endTime) : null;
                
                workElement.innerHTML = `
                    <div class="d-flex w-100 justify-content-between">
                        <h5 class="mb-1">${work.employeeName}</h5>
                        <small>${startTime.toLocaleTimeString('tr-TR')}</small>
                    </div>
                    <p class="mb-1">${work.workType === 'software' ? 'Yazılım' : 'Video'} İşi</p>
                    ${work.videoLink ? `<small>Video: <a href="${work.videoLink}" target="_blank">${work.videoLink}</a></small><br>` : ''}
                    <small class="text-muted">Durum: ${work.status === 'completed' ? 
                        `Tamamlandı (${endTime.toLocaleTimeString('tr-TR')})` : 
                        'Devam Ediyor'}</small>
                    ${work.status === 'in_progress' ? 
                        `<button class="btn btn-success btn-sm mt-2" onclick="completeWork('${work.id}', '${work.workType}')">İşi Tamamla</button>` : 
                        ''}
                `;
                activeWorksContainer.appendChild(workElement);
            });
        } catch (error) {
            console.error('Error:', error);
        }
    }

    // Initialize the page
    loadEmployees();
    loadActiveWorks();
    
    // Refresh works every minute
    setInterval(loadActiveWorks, 60000);
});

// Complete work function
async function completeWork(workId, workType) {
    const modal = new bootstrap.Modal(document.getElementById('completeWorkModal'));
    document.getElementById('completeWorkId').value = workId;
    
    // Show/hide video link field based on work type
    const videoLinkField = document.getElementById('videoLinkField');
    const videoLinkInput = document.getElementById('videoLink');
    if (workType === 'video') {
        videoLinkField.style.display = 'block';
        videoLinkInput.required = true;
    } else {
        videoLinkField.style.display = 'none';
        videoLinkInput.required = false;
    }
    
    modal.show();
}

// Handle work completion
document.getElementById('saveComplete').addEventListener('click', async function() {
    const workId = document.getElementById('completeWorkId').value;
    const videoLink = document.getElementById('videoLink').value;

    if (document.getElementById('videoLinkField').style.display === 'block' && !videoLink) {
        alert('Lütfen video linkini girin!');
        return;
    }

    try {
        const response = await fetch(`/api/work/${workId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                endTime: new Date().toISOString(),
                videoLink: videoLink || null
            })
        });

        if (!response.ok) {
            throw new Error('Network response was not ok');
        }

        const modal = bootstrap.Modal.getInstance(document.getElementById('completeWorkModal'));
        modal.hide();
        document.getElementById('videoLink').value = '';
        loadActiveWorks();
        alert('İş başarıyla tamamlandı!');
    } catch (error) {
        console.error('Error:', error);
        alert('İş tamamlanırken bir hata oluştu!');
    }
}); 