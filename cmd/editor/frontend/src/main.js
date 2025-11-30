import {
  SaveStory,
  GetImageList,
  GetSoundList,
  GetSpriteList,
  LoadStory,
} from "../wailsjs/go/main/App";
import Sortable from "sortablejs";

const listElement = document.getElementById("list");
const editModal = document.getElementById("edit-modal");
const confirmModal = document.getElementById("confirm-modal");

let currentEditingCard = null;
let pendingDeleteTarget = null;
let currentLang = "en"; // 현재 언어 상태

// =========================================================
// 🌐 번역 데이터 (Dictionary)
// =========================================================
const translations = {
  en: {
    // UI Text
    tab_action: "🎬 Action",
    tab_config: "⚙️ Config",
    menu_actions: "ACTIONS",
    btn_dialogue: "Add Dialogue",
    btn_video: "Add Video",
    btn_branch: "Add Branch",
    lbl_title: "Game Title",
    menu_editor: "EDITOR PREFERENCES",
    lbl_language: "Language",
    lbl_theme: "Theme",
    btn_save: "Save Project",

    // Modal & Dynamic Text
    modal_edit_dialogue: "✏️ Edit Dialogue",
    modal_edit_video: "🎬 Edit Video",
    ph_actor: "Character Name",
    ph_text: "Enter dialogue here...",
    ph_cond: "e.g. love >= 100",
    msg_delete_confirm: "Delete?",
    msg_delete_desc: "Are you sure you want to delete this?",
    msg_delete_warn: "This action cannot be undone.",
    btn_cancel: "Cancel",
    btn_apply: "Apply",
    btn_delete: "Delete",

    // Toasts
    msg_load_success: "Project Loaded Successfully ✨",
    msg_save_success: "Project Saved Successfully 💾",
    msg_saved_mod: "Modified Successfully",
    msg_deleted: "Item Deleted",
  },
  ko: {
    // UI Text
    tab_action: "🎬 제작",
    tab_config: "⚙️ 설정",
    menu_actions: "액션 추가",
    btn_dialogue: "대사 추가",
    btn_video: "영상 연출",
    btn_branch: "분기점 생성",
    lbl_title: "게임 제목",
    menu_editor: "에디터 환경설정",
    lbl_language: "언어 (Language)",
    lbl_theme: "테마 (Theme)",
    btn_save: "프로젝트 저장",

    // Modal & Dynamic Text
    modal_edit_dialogue: "✏️ 대사 편집",
    modal_edit_video: "🎬 영상 설정",
    ph_actor: "캐릭터 이름",
    ph_text: "대사를 입력하세요...",
    ph_cond: "예: love >= 100",
    msg_delete_confirm: "삭제 확인",
    msg_delete_desc: "정말 이 카드를 삭제하시겠습니까?",
    msg_delete_warn: "이 작업은 되돌릴 수 없습니다.",
    btn_cancel: "취소",
    btn_apply: "적용",
    btn_delete: "삭제",

    // Toasts
    msg_load_success: "프로젝트를 불러왔습니다 ✨",
    msg_save_success: "저장되었습니다 💾",
    msg_saved_mod: "수정되었습니다",
    msg_deleted: "삭제되었습니다",
  },
};

// =========================================================
// 0. 초기화 (Init)
// =========================================================
(async function init() {
  // 1. 드래그 앤 드롭 활성화 (이게 없으면 위치 변경 불가)
  new Sortable(listElement, {
    animation: 150,
    handle: ".script-card",
    ghostClass: "sortable-ghost",
  });

  // 2. 테마 복구
  const savedTheme = localStorage.getItem("rengui_theme");
  if (savedTheme === "light") {
    document.body.classList.add("light-mode");
    document.getElementById("theme-label").innerText = "☀️ Light Mode";
  }

  // 3. 언어 복구
  const savedLang = localStorage.getItem("rengui_lang") || "en";
  document.getElementById("app-lang").value = savedLang;
  changeLanguage(savedLang);

  // 4. 데이터 로드
  try {
    const jsonStr = await LoadStory();
    if (jsonStr) {
      const data = JSON.parse(jsonStr);
      // 설정 탭 복구
      if (data.system) {
        document.getElementById("conf-title").value =
          data.system.title || "RenGui Game";
        document.getElementById("conf-w").value =
          data.system.screenWidth || 1280;
        document.getElementById("conf-h").value =
          data.system.screenHeight || 720;
      }
      if (data.ui) {
        document.getElementById("conf-box-color").value =
          data.ui.boxColor || "#000000";
        document.getElementById("conf-box-opacity").value =
          data.ui.boxOpacity || 0.7;
        document.getElementById("conf-text-color").value =
          data.ui.textColor || "#FFFFFF";
        document.getElementById("conf-box-height").value =
          data.ui.boxHeight || 200;
      }

      // 카드 생성
      if (data.scenes && data.scenes.length > 0) {
        document.getElementById("list").innerHTML = ""; // 초기화
        data.scenes[0].dialogues.forEach((d) => {
          let type = d.video ? "video" : "dialogue";
          if (d.choices && d.choices.length > 0) type = "branch";
          createCardUI(type, d);
        });
        showToast(translations[currentLang]["msg_load_success"], "success");
      }
    }
  } catch (e) {
    console.log("New Project Mode");
  }
})();

// =========================================================
// 1. 카드 생성 및 UI
// =========================================================
window.addCard = function (type) {
  let data = {
    type: type,
    actor: type === "dialogue" ? "New Actor" : "",
    text: type === "dialogue" ? "..." : "",
    background: "",
    video: "",
    bgm: "",
    sfx: "",
    condition: "",
  };

  if (type === "video") {
    data.actor = "Video Clip";
    data.text = "(No file selected)";
  }

  createCardUI(type, data);

  setTimeout(() => {
    document.querySelector(".main").scrollTo(0, document.body.scrollHeight);
    // 생성 직후 수정창 띄우기
    const last = listElement.lastElementChild;
    const editBtn = last.querySelector(".btn-edit-trigger");
    if (editBtn) openEdit(editBtn);
  }, 100);
};

function createCardUI(type, data) {
  const li = document.createElement("li");
  li.className = `script-card type-${type}`;
  li.dataset.json = JSON.stringify(data);

  let icon = "comment";
  if (type === "video") icon = "film";
  if (type === "branch") icon = "code-branch";

  // ★ onclick 이벤트를 직접 HTML에 박아넣어서 확실하게 작동시킴
  li.innerHTML = `
        <div class="card-icon"><i class="fas fa-${icon}"></i></div>
        <div class="card-content">
            ${
              data.actor ? `<span class="actor-badge">${data.actor}</span>` : ""
            }
            <div class="dialogue-text">${data.text || data.video || "..."}</div>
            <div class="meta-info">
                ${
                  data.background
                    ? `<span class="meta-tag"><i class="fas fa-image"></i> ${data.background}</span>`
                    : ""
                }
                ${
                  data.bgm
                    ? `<span class="meta-tag"><i class="fas fa-music"></i> ${data.bgm}</span>`
                    : ""
                }
            </div>
        </div>
        <div class="card-actions">
            <button class="btn-icon btn-edit-trigger" onclick="openEdit(this)">
                <i class="fas fa-pen"></i>
            </button>
            <button class="btn-icon btn-delete-trigger" onclick="reqDeleteCard(this)">
                <i class="fas fa-trash"></i>
            </button>
        </div>
    `;
  listElement.appendChild(li);
}

// =========================================================
// 2. 수정 (Edit)
// =========================================================
window.openEdit = async function (btn) {
  // 버튼에서 가장 가까운 카드 찾기 (이게 핵심)
  const li = btn.closest(".script-card");
  currentEditingCard = li;

  const data = JSON.parse(li.dataset.json);
  const type = data.type || (data.video ? "video" : "dialogue");

  // ★ 번역된 제목 적용
  const titleKey =
    type === "video" ? "modal_edit_video" : "modal_edit_dialogue";
  document.getElementById("modal-title").innerText =
    translations[currentLang][titleKey];

  const uiDialogue = document.getElementById("field-dialogue");
  const uiVideo = document.getElementById("field-video");

  if (type === "video") {
    uiDialogue.style.display = "none";
    uiVideo.style.display = "block";

    const files = (await GetImageList()) || [];
    const sel = document.getElementById("edit-video");
    sel.innerHTML = '<option value="">(Select)</option>';
    files.forEach((f) => {
      if (f.endsWith(".ivf"))
        sel.innerHTML += `<option value="${f}" ${
          f === data.video ? "selected" : ""
        }>${f}</option>`;
    });
  } else {
    uiDialogue.style.display = "block";
    uiVideo.style.display = "none";

    document.getElementById("edit-actor").value = data.actor || "";
    document.getElementById("edit-text").value = data.text || "";

    // 배경 목록
    const files = (await GetImageList()) || [];
    const sel = document.getElementById("edit-bg");
    sel.innerHTML = '<option value="">(None)</option>';
    files.forEach((f) => {
      if (!f.endsWith(".ivf"))
        sel.innerHTML += `<option value="${f}" ${
          f === data.background ? "selected" : ""
        }>${f}</option>`;
    });

    // 캐릭터 목록
    const sprites = (await GetSpriteList()) || [];
    const fillOptions = (id, val) => {
      const el = document.getElementById(id);
      el.innerHTML = '<option value="">(None)</option>';
      sprites.forEach(
        (s) =>
          (el.innerHTML += `<option value="${s}" ${
            s === val ? "selected" : ""
          }>${s}</option>`)
      );
    };
    fillOptions("edit-char-left", data.charLeft);
    fillOptions("edit-char-center", data.charCenter);
    fillOptions("edit-char-right", data.charRight);
  }

  // 공통: 사운드
  const sounds = (await GetSoundList()) || [];
  const bgmSel = document.getElementById("edit-bgm");
  const sfxSel = document.getElementById("edit-sfx");

  bgmSel.innerHTML = '<option value="">(Keep)</option>';
  sfxSel.innerHTML = '<option value="">(None)</option>';

  sounds.forEach((s) => {
    bgmSel.innerHTML += `<option value="${s}" ${
      s === data.bgm ? "selected" : ""
    }>${s}</option>`;
    sfxSel.innerHTML += `<option value="${s}" ${
      s === data.sfx ? "selected" : ""
    }>${s}</option>`;
  });

  document.getElementById("edit-cond").value = data.condition || "";

  editModal.showModal();
};

window.confirmEdit = function () {
  if (!currentEditingCard) return;

  let data = JSON.parse(currentEditingCard.dataset.json);
  const type = data.type || (data.video ? "video" : "dialogue");

  if (type === "video") {
    data.video = document.getElementById("edit-video").value;
    data.text = data.video;
    data.actor = "Video Clip";
  } else {
    data.actor = document.getElementById("edit-actor").value;
    data.text = document.getElementById("edit-text").value;
    data.background = document.getElementById("edit-bg").value;
    data.charLeft = document.getElementById("edit-char-left").value;
    data.charCenter = document.getElementById("edit-char-center").value;
    data.charRight = document.getElementById("edit-char-right").value;
  }

  data.bgm = document.getElementById("edit-bgm").value;
  data.sfx = document.getElementById("edit-sfx").value;
  data.condition = document.getElementById("edit-cond").value;

  currentEditingCard.dataset.json = JSON.stringify(data);

  // UI 갱신 (HTML 다시 그리기)
  const contentDiv = currentEditingCard.querySelector(".card-content");
  let metaHTML = `
        ${
          data.background
            ? `<span class="meta-tag"><i class="fas fa-image"></i> ${data.background}</span>`
            : ""
        }
        ${
          data.bgm
            ? `<span class="meta-tag"><i class="fas fa-music"></i> ${data.bgm}</span>`
            : ""
        }
        ${
          data.condition
            ? `<span class="meta-tag" style="color:#e67e22"><i class="fas fa-question-circle"></i> ${data.condition}</span>`
            : ""
        }
    `;

  contentDiv.innerHTML = `
        ${data.actor ? `<span class="actor-badge">${data.actor}</span>` : ""}
        <div class="dialogue-text">${data.text || "..."}</div>
        <div class="meta-info">${metaHTML}</div>
    `;

  closeModal();
  showToast(translations[currentLang]["msg_saved_mod"], "success");
};

// =========================================================
// 3. 삭제 (Delete)
// =========================================================
window.reqDeleteCard = function (btn) {
  pendingDeleteTarget = btn.closest(".script-card");
  confirmModal.showModal();
};

window.closeConfirm = function (isYes) {
  confirmModal.close();
  if (isYes && pendingDeleteTarget) {
    pendingDeleteTarget.remove();
    showToast(translations[currentLang]["msg_deleted"], "error");
  }
  pendingDeleteTarget = null;
};

window.closeModal = () => editModal.close();

// =========================================================
// 4. 저장 및 설정 (Save & Config)
// =========================================================
window.saveFile = function () {
  const cards = document.querySelectorAll(".script-card");
  let dialogues = [];
  cards.forEach((c) => dialogues.push(JSON.parse(c.dataset.json)));

  const gameData = {
    version: 1,
    system: {
      title: document.getElementById("conf-title").value,
      screenWidth: parseInt(document.getElementById("conf-w").value),
      screenHeight: parseInt(document.getElementById("conf-h").value),
    },
    ui: {
      boxColor: document.getElementById("conf-box-color").value,
      boxOpacity: parseFloat(document.getElementById("conf-box-opacity").value),
      textColor: document.getElementById("conf-text-color").value,
      boxHeight: parseInt(document.getElementById("conf-box-height").value),
      fontSize: 24,
    },
    variables: {},
    scenes: [{ id: "scene_01", dialogues: dialogues }],
  };

  SaveStory(JSON.stringify(gameData)).then((res) => {
    // 성공 메시지도 번역해서 보여줌
    showToast(translations[currentLang]["msg_save_success"], "success");
  });
};

// 언어 변경 함수
window.changeLanguage = function (lang) {
  currentLang = lang;
  localStorage.setItem("rengui_lang", lang);

  // 1. data-i18n 태그 텍스트 교체
  document.querySelectorAll("[data-i18n]").forEach((el) => {
    const key = el.dataset.i18n;
    if (translations[lang][key]) {
      el.innerText = translations[lang][key];
    }
  });

  // 2. Input Placeholder 교체
  const setPlaceHolder = (id, key) => {
    const el = document.getElementById(id);
    if (el) el.placeholder = translations[lang][key];
  };
  setPlaceHolder("edit-actor", "ph_actor");
  setPlaceHolder("edit-text", "ph_text");
  setPlaceHolder("edit-cond", "ph_cond");
};

// 테마 변경 함수
window.toggleTheme = function () {
  const body = document.body;
  const label = document.getElementById("theme-label");

  if (body.classList.contains("light-mode")) {
    body.classList.remove("light-mode");
    label.innerText = "🌙 Dark Mode";
    localStorage.setItem("rengui_theme", "dark");
  } else {
    body.classList.add("light-mode");
    label.innerText = "☀️ Light Mode";
    localStorage.setItem("rengui_theme", "light");
  }
};

window.switchTab = function (tabName) {
  const actionTab = document.getElementById("tab-action");
  const configTab = document.getElementById("tab-config");
  const btns = document.querySelectorAll(".tab-btn");

  // 스타일 초기화 로직 (단순화)
  btns.forEach((b) => {
    b.classList.remove("active");
    b.style.background = "transparent";
    b.style.color = "var(--text-muted)";
  });

  // 선택된 탭 활성화
  const activeBtn = tabName === "action" ? btns[0] : btns[1];
  activeBtn.classList.add("active");
  activeBtn.style.background = "var(--accent-gradient)";
  activeBtn.style.color = "white";

  actionTab.style.display = tabName === "action" ? "block" : "none";
  configTab.style.display = tabName === "config" ? "block" : "none";
};

window.showToast = function (msg, type = "info") {
  const container = document.getElementById("toast-container");
  const toast = document.createElement("div");
  toast.className = "toast";

  let icon =
    type === "success"
      ? '<i class="fas fa-check-circle"></i> '
      : '<i class="fas fa-info-circle"></i> ';
  if (type === "error") icon = '<i class="fas fa-trash-alt"></i> ';

  toast.innerHTML = icon + msg;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = "0";
    toast.style.transform = "translateY(10px)";
    setTimeout(() => toast.remove(), 300);
  }, 3000);
};
