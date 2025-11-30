import {SaveStory, GetImageList, GetSoundList, GetSpriteList, LoadStory} from '../wailsjs/go/main/App';
import Sortable from 'sortablejs';

const listElement = document.getElementById('list');
const editModal = document.getElementById('edit-modal');
const confirmModal = document.getElementById('confirm-modal');

let currentEditingCard = null;
let pendingDeleteTarget = null;

// =========================================================
// 0. 초기화 및 이벤트 리스너 (여기가 핵심!)
// =========================================================
(async function init() {
    // 1. 드래그 앤 드롭 설정
    new Sortable(listElement, {
        animation: 150,
        handle: '.script-card', 
        ghostClass: 'sortable-ghost'
    });

    // 2. ★ 클릭 이벤트 위임 (가장 중요한 부분)
    // HTML에 onclick을 쓰지 않고 여기서 처리합니다.
    document.addEventListener('click', function(e) {
        
        // (1) 수정(연필) 버튼 클릭 감지
        const editBtn = e.target.closest('.btn-edit-trigger');
        if (editBtn) {
            openEdit(editBtn); // 함수 호출
            return;
        }

        // (2) 삭제(휴지통) 버튼 클릭 감지
        const delBtn = e.target.closest('.btn-delete-trigger');
        if (delBtn) {
            reqDeleteCard(delBtn); // 함수 호출
            return;
        }
    });

    // 3. 데이터 로드
    try {
        const jsonStr = await LoadStory();
        if(jsonStr) {
            const data = JSON.parse(jsonStr);
            // 설정값 불러오기 (없으면 기본값)
            if(data.system) {
                document.getElementById('conf-title').value = data.system.title || "My Game";
                document.getElementById('conf-w').value = data.system.screenWidth || 1280;
                document.getElementById('conf-h').value = data.system.screenHeight || 720;
            }
            if(data.ui) {
                document.getElementById('conf-box-color').value = data.ui.boxColor || "#000000";
                document.getElementById('conf-box-opacity').value = data.ui.boxOpacity || 0.7;
                document.getElementById('conf-text-color').value = data.ui.textColor || "#FFFFFF";
                document.getElementById('conf-box-height').value = data.ui.boxHeight || 200;
            }
            if(data.scenes && data.scenes.length > 0) {
                data.scenes[0].dialogues.forEach(d => {
                    let type = d.video ? 'video' : 'dialogue';
                    if (d.choices && d.choices.length > 0) type = 'branch';
                    createCardUI(type, d);
                });
                showToast("프로젝트 로드 완료");
            }
        }
    } catch(e) { console.log("새 프로젝트 모드"); }
})();

// =========================================================
// 1. 카드 생성 (onclick 제거됨)
// =========================================================
window.addCard = function(type) {
    let data = {
        type: type,
        actor: type === 'dialogue' ? 'New Actor' : '',
        text: type === 'dialogue' ? '내용을 입력하세요' : '',
        background: '', video: '', bgm: '', sfx: '', condition: ''
    };
    
    if (type === 'video') {
        data.actor = "영상 연출";
        data.text = "(영상 파일 미선택)";
    }

    createCardUI(type, data);
    
    setTimeout(() => {
        document.querySelector('.main').scrollTo(0, document.body.scrollHeight);
        // 생성 후 바로 수정창 띄우기
        const last = listElement.lastElementChild;
        const editBtn = last.querySelector('.btn-edit-trigger'); 
        if(editBtn) openEdit(editBtn);
    }, 100);
};

function createCardUI(type, data) {
    const li = document.createElement('li');
    li.className = `script-card type-${type}`;
    li.dataset.json = JSON.stringify(data);

    let icon = 'comment';
    if(type === 'video') icon = 'film';
    if(type === 'branch') icon = 'code-branch';

    // ★ HTML 문자열에서 onclick을 제거하고 class만 부여합니다.
    li.innerHTML = `
        <div class="card-icon"><i class="fas fa-${icon}"></i></div>
        <div class="card-content">
            ${data.actor ? `<span class="actor-badge">${data.actor}</span>` : ''}
            <div class="dialogue-text">${data.text || data.video || "..."}</div>
            <div class="meta-info">
                ${data.background ? `<span class="meta-tag"><i class="fas fa-image"></i> ${data.background}</span>` : ''}
                ${data.bgm ? `<span class="meta-tag"><i class="fas fa-music"></i> ${data.bgm}</span>` : ''}
            </div>
        </div>
        <div class="card-actions">
            <button class="btn-icon btn-edit-trigger">
                <i class="fas fa-pen"></i>
            </button>
            <button class="btn-icon btn-delete-trigger">
                <i class="fas fa-trash"></i>
            </button>
        </div>
    `;
    listElement.appendChild(li);
}

// =========================================================
// 2. 수정 로직 (e.stopPropagation 불필요)
// =========================================================
async function openEdit(btn) {
    const li = btn.closest('.script-card');
    currentEditingCard = li;
    
    const data = JSON.parse(li.dataset.json);
    const type = data.type || (data.video ? 'video' : 'dialogue');

    // UI 세팅
    document.getElementById('modal-title').innerText = type === 'video' ? '🎬 영상 설정' : '✏️ 대사 편집';
    const uiDialogue = document.getElementById('field-dialogue');
    const uiVideo = document.getElementById('field-video');

    if (type === 'video') {
        uiDialogue.style.display = 'none';
        uiVideo.style.display = 'block';
        
        // ★ 수정 1: || [] 추가 (파일 없으면 빈 배열로)
        const files = await GetImageList() || []; 
        const sel = document.getElementById('edit-video');
        sel.innerHTML = '<option value="">(선택)</option>';
        files.forEach(f => {
            if(f.endsWith('.ivf')) sel.innerHTML += `<option value="${f}" ${f===data.video?'selected':''}>${f}</option>`;
        });
    } else {
        uiDialogue.style.display = 'block';
        uiVideo.style.display = 'none';
        
        document.getElementById('edit-actor').value = data.actor || "";
        document.getElementById('edit-text').value = data.text || "";
        
        // ★ [신규] 캐릭터 목록 로드 및 선택
        const sprites = await GetSpriteList() || [];
        const leftSel = document.getElementById('edit-char-left');
        const centerSel = document.getElementById('edit-char-center');
        const rightSel = document.getElementById('edit-char-right');
        
        // 헬퍼: 옵션 초기화 및 목록 채우기
        const fillOptions = (sel, currentVal) => {
            sel.innerHTML = '<option value="">(비움)</option>';
            sprites.forEach(f => {
                sel.innerHTML += `<option value="${f}" ${f===currentVal?'selected':''}>${f}</option>`;
            });
        };

        fillOptions(leftSel, data.charLeft);
        fillOptions(centerSel, data.charCenter);
        fillOptions(rightSel, data.charRight);


        // 배경 이미지 목록 (기존 코드 유지)
        const files = await GetImageList() || [];
        const sel = document.getElementById('edit-bg');
        sel.innerHTML = '<option value="">(없음)</option>';
        files.forEach(f => {
            if(!f.endsWith('.ivf')) sel.innerHTML += `<option value="${f}" ${f===data.background?'selected':''}>${f}</option>`;
        });
    }

    // ★ 수정 3: || [] 추가 (소리가 없어도 에러 안 나게)
    const sounds = await GetSoundList() || [];
    const bgmSel = document.getElementById('edit-bgm');
    const sfxSel = document.getElementById('edit-sfx');
    
    bgmSel.innerHTML = '<option value="">(유지)</option>';
    sfxSel.innerHTML = '<option value="">(없음)</option>';
    
    sounds.forEach(s => {
        bgmSel.innerHTML += `<option value="${s}" ${s===data.bgm?'selected':''}>${s}</option>`;
        sfxSel.innerHTML += `<option value="${s}" ${s===data.sfx?'selected':''}>${s}</option>`;
    });

    document.getElementById('edit-cond').value = data.condition || "";

    editModal.showModal();
}

// 저장 버튼 (HTML onclick 연결용 window 함수)
window.confirmEdit = function() {
    if(!currentEditingCard) return;
    
    let data = JSON.parse(currentEditingCard.dataset.json);
    const type = data.type || (data.video ? 'video' : 'dialogue');

    if(type === 'video') {
        data.video = document.getElementById('edit-video').value;
        data.text = data.video;
        data.actor = "영상 연출";
    } else {
        data.actor = document.getElementById('edit-actor').value;
        data.text = document.getElementById('edit-text').value;
        data.background = document.getElementById('edit-bg').value;
        
        data.charLeft = document.getElementById('edit-char-left').value;
        data.charCenter = document.getElementById('edit-char-center').value;
        data.charRight = document.getElementById('edit-char-right').value;
    }

    data.bgm = document.getElementById('edit-bgm').value;
    data.sfx = document.getElementById('edit-sfx').value;
    data.condition = document.getElementById('edit-cond').value;

    currentEditingCard.dataset.json = JSON.stringify(data);
    
    // UI 갱신
    const contentDiv = currentEditingCard.querySelector('.card-content');
    
    let metaHTML = `
        ${data.background ? `<span class="meta-tag"><i class="fas fa-image"></i> ${data.background}</span>` : ''}
        ${data.bgm ? `<span class="meta-tag"><i class="fas fa-music"></i> ${data.bgm}</span>` : ''}
        ${data.condition ? `<span class="meta-tag" style="color:#e67e22"><i class="fas fa-question-circle"></i> ${data.condition}</span>` : ''}
    `;

    contentDiv.innerHTML = `
        ${data.actor ? `<span class="actor-badge">${data.actor}</span>` : ''}
        <div class="dialogue-text">${data.text || "..."}</div>
        <div class="meta-info">${metaHTML}</div>
    `;

    closeModal();
    showToast("수정되었습니다", "success");
};

// =========================================================
// 3. 삭제 및 기타 (window 함수)
// =========================================================
window.closeModal = () => editModal.close();

// 삭제 요청 (이벤트 위임에서 호출됨)
function reqDeleteCard(btn) {
    pendingDeleteTarget = btn.closest('.script-card');
    confirmModal.showModal();
}

window.closeConfirm = function(isYes) {
    confirmModal.close();
    if (isYes && pendingDeleteTarget) {
        pendingDeleteTarget.remove();
        showToast("삭제되었습니다", "error");
    }
    pendingDeleteTarget = null;
};

window.saveFile = function() {
    const cards = document.querySelectorAll('.script-card');
    let dialogues = [];
    cards.forEach(c => dialogues.push(JSON.parse(c.dataset.json)));

    const gameData = {
        version: 1,
        system: {
            title: document.getElementById('conf-title').value,
            screenWidth: parseInt(document.getElementById('conf-w').value),
            screenHeight: parseInt(document.getElementById('conf-h').value),
        },
        ui: {
            boxColor: document.getElementById('conf-box-color').value,
            boxOpacity: parseFloat(document.getElementById('conf-box-opacity').value),
            textColor: document.getElementById('conf-text-color').value,
            boxHeight: parseInt(document.getElementById('conf-box-height').value),
            fontSize: 24
        },
        scenes: [{ id: "scene_01", dialogues: dialogues }]
    };

    SaveStory(JSON.stringify(gameData)).then(res => showToast(res, "success"));
};

window.showToast = function(msg, type='info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = 'toast';
    
    if(type === 'success') toast.style.borderLeft = "5px solid #2ecc71";
    if(type === 'error') toast.style.borderLeft = "5px solid #e74c3c";
    
    let icon = type==='success' ? '<i class="fas fa-check-circle"></i> ' : '<i class="fas fa-info-circle"></i> ';
    if(type==='error') icon = '<i class="fas fa-trash-alt"></i> ';
    
    toast.innerHTML = icon + msg;
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(10px)';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
};

// 탭 전환 (버튼용)
window.switchTab = function(tabName) {
    const actionTab = document.getElementById('tab-action');
    const configTab = document.getElementById('tab-config');
    const btns = document.querySelectorAll('.tab-btn');

    if(tabName === 'action') {
        actionTab.style.display = 'block';
        configTab.style.display = 'none';
        btns[0].style.background = '#4285F4'; btns[0].style.color = 'white';
        btns[1].style.background = '#2D2E35'; btns[1].style.color = '#aaa';
    } else {
        actionTab.style.display = 'none';
        configTab.style.display = 'block';
        btns[0].style.background = '#2D2E35'; btns[0].style.color = '#aaa';
        btns[1].style.background = '#4285F4'; btns[1].style.color = 'white';
    }
};