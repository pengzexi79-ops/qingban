const { createApp, ref, computed, watch, onMounted, nextTick } = Vue;

const app = createApp({
  setup() {
    const currentPage = ref('chatlist');
    const searchQuery = ref('');
    const showChatMenu = ref(false);
    const companions = ref(Store.getCompanions());
    const currentChatId = ref(null);
    const inputMessage = ref('');
    const messageList = ref(null);
    const editingCompanion = ref(null);
    const userSettings = ref(Store.getSettings());
    const pageHistory = ref([]);

    const avatarColors = [
      '#FF6B6B', '#FF8E53', '#FFC857', '#57C47A',
      '#4ECDC4', '#45B7D1', '#6C5CE7', '#A855F7',
      '#EC4899', '#8B5CF6', '#06B6D4', '#10B981'
    ];

    const companionForm = ref(getEmptyForm());

    function getEmptyForm() {
      return {
        name: '', avatarColor: '#57C47A', personality: '',
        roleSetting: '', greeting: '', canAutoMessage: true
      };
    }

    const pageTitle = computed(() => {
      const titles = { chatlist: '青伴', companions: 'AI伴侣', settings: '设置' };
      return titles[currentPage.value] || '青伴';
    });

    const filteredCompanions = computed(() => {
      if (!searchQuery.value) return companions.value;
      const q = searchQuery.value.toLowerCase();
      return companions.value.filter(c =>
        c.name.toLowerCase().includes(q) ||
        (c.personality && c.personality.toLowerCase().includes(q))
      );
    });

    const currentCompanion = computed(() => {
      return companions.value.find(c => c.id === currentChatId.value);
    });

    const currentChatName = computed(() => {
      return currentCompanion.value ? currentCompanion.value.name : '';
    });

    const currentMessages = computed(() => {
      if (!currentChatId.value) return [];
      return Store.getChatMessages(currentChatId.value);
    });

    function goAddCompanion() {
      editingCompanion.value = null;
      companionForm.value = getEmptyForm();
      pageHistory.value.push(currentPage.value);
      currentPage.value = 'addCompanion';
    }

    function goBack() {
      const prev = pageHistory.value.pop();
      currentPage.value = prev || 'chatlist';
    }

    function openChat(companion) {
      currentChatId.value = companion.id;
      pageHistory.value.push(currentPage.value);
      currentPage.value = 'chat';
      const messages = Store.getChatMessages(companion.id);
      if (messages.length === 0 && companion.greeting) {
        Store.addMessage(companion.id, { role: 'assistant', content: companion.greeting });
      }
      nextTick(() => { scrollToBottom(); });
    }

    function sendMessage() {
      const text = inputMessage.value.trim();
      if (!text || !currentChatId.value) return;
      Store.addMessage(currentChatId.value, { role: 'user', content: text });
      inputMessage.value = '';
      companions.value = [...companions.value];
      nextTick(() => { scrollToBottom(); });

      setTimeout(() => {
        const reply = generateReply(text, currentCompanion.value);
        Store.addMessage(currentChatId.value, { role: 'assistant', content: reply });
        companions.value = [...companions.value];
        nextTick(() => { scrollToBottom(); });
      }, 500 + Math.random() * 1000);
    }

    function scrollToBottom() {
      if (messageList.value) {
        messageList.value.scrollTop = messageList.value.scrollHeight;
      }
    }

    function saveCompanion() {
      const form = companionForm.value;
      if (!form.name.trim()) { alert('请输入伴侣名字'); return; }
      if (editingCompanion.value) {
        Store.updateCompanion(editingCompanion.value.id, {
          name: form.name, avatarColor: form.avatarColor,
          personality: form.personality, roleSetting: form.roleSetting,
          greeting: form.greeting, canAutoMessage: form.canAutoMessage
        });
      } else {
        const nc = Store.addCompanion({
          name: form.name, avatarColor: form.avatarColor,
          personality: form.personality, roleSetting: form.roleSetting,
          greeting: form.greeting, canAutoMessage: form.canAutoMessage
        });
        if (form.greeting) {
          Store.addMessage(nc.id, { role: 'assistant', content: form.greeting });
        }
      }
      companions.value = Store.getCompanions();
      goBack();
    }

    function editCompanion(companion) {
      editingCompanion.value = companion;
      companionForm.value = {
        name: companion.name, avatarColor: companion.avatarColor,
        personality: companion.personality || '',
        roleSetting: companion.roleSetting || '',
        greeting: companion.greeting || '',
        canAutoMessage: companion.canAutoMessage !== false
      };
      pageHistory.value.push(currentPage.value);
      currentPage.value = 'addCompanion';
    }

    function deleteCompanion(id) {
      if (confirm('确定要删除这个AI伴侣吗？所有聊天记录将被清除。')) {
        Store.deleteCompanion(id);
        companions.value = Store.getCompanions();
      }
    }

    function getLastMessage(companionId) {
      const messages = Store.getChatMessages(companionId);
      if (messages.length === 0) return '暂无消息';
      const last = messages[messages.length - 1];
      return (last.role === 'user' ? '[我] ' : '') + last.content;
    }

    function getLastTime(companionId) {
      const messages = Store.getChatMessages(companionId);
      if (messages.length === 0) return '';
      return formatTime(messages[messages.length - 1].timestamp);
    }

    function formatTime(timestamp) {
      const date = new Date(timestamp);
      const now = new Date();
      if (date.toDateString() === now.toDateString()) {
        return date.getHours().toString().padStart(2, '0') + ':' + date.getMinutes().toString().padStart(2, '0');
      }
      const yesterday = new Date(now);
      yesterday.setDate(yesterday.getDate() - 1);
      if (date.toDateString() === yesterday.toDateString()) return '昨天';
      return (date.getMonth() + 1) + '/' + date.getDate();
    }

    function exportMemory() {
      const data = Store.exportData();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'qingban_backup_' + new Date().toISOString().slice(0, 10) + '.json';
      a.click();
      URL.revokeObjectURL(url);
    }

    function clearAllData() {
      if (confirm('确定要清除所有数据吗？此操作不可恢复。')) {
        Store.clearAll();
        companions.value = [];
        userSettings.value = Store.getSettings();
      }
    }

    function generateReply(userMsg, companion) {
      if (!companion) return '...';
      const p = companion.personality || '';
      const warm = ['抱抱你~', '你真的很棒！', '我相信你可以的！', '无论发生什么，我都会在你身边。', '你对我来说很重要。'];
      const calm = ['我理解。', '这确实值得思考。', '你的想法很有道理。', '让我们一起想想办法。', '慢慢来，不急。'];
      const def = ['我在听呢，继续说~', '嗯嗯，我理解你的感受。', '谢谢你愿意和我分享这些。', '你今天过得怎么样？', '我一直在这里陪着你哦。'];
      let pool = def;
      if (p.includes('温柔') || p.includes('甜') || p.includes('暖')) pool = warm;
      else if (p.includes('冷静') || p.includes('理性') || p.includes('沉稳')) pool = calm;
      if (userMsg.includes('你好') || userMsg.includes('hi')) return '你好呀！很高兴见到你~';
      if (userMsg.includes('再见') || userMsg.includes('bye')) return '下次见哦，我会想你的~';
      if (userMsg.includes('难过') || userMsg.includes('伤心')) return '抱抱你，会好起来的。我在这里陪你。';
      if (userMsg.includes('开心') || userMsg.includes('高兴')) return '看到你开心，我也很高兴呢！';
      return pool[Math.floor(Math.random() * pool.length)];
    }

    let autoMessageTimer = null;
    function startAutoMessages() {
      const scheduleNext = () => {
        const delay = 30000 + Math.random() * 90000;
        autoMessageTimer = setTimeout(() => {
          if (userSettings.value.autoMsgNotify) {
            const eligible = companions.value.filter(c => c.canAutoMessage !== false);
            if (eligible.length > 0) {
              const comp = eligible[Math.floor(Math.random() * eligible.length)];
              const msgs = ['在吗？想你了~', '今天过得怎么样？', '有空聊聊天吗？', '刚刚想到你了~', '你在忙什么呢？'];
              Store.addMessage(comp.id, { role: 'assistant', content: msgs[Math.floor(Math.random() * msgs.length)] });
              companions.value = [...companions.value];
            }
          }
          scheduleNext();
        }, delay);
      };
      scheduleNext();
    }

    watch(userSettings, (val) => { Store.saveSettings(val); }, { deep: true });

    onMounted(() => { startAutoMessages(); });

    return {
      currentPage, searchQuery, showChatMenu, companions, currentChatId,
      inputMessage, messageList, editingCompanion, userSettings, companionForm, avatarColors,
      pageTitle, filteredCompanions, currentCompanion, currentChatName, currentMessages,
      goAddCompanion, goBack, openChat, sendMessage, saveCompanion, editCompanion,
      deleteCompanion, getLastMessage, getLastTime, formatTime, exportMemory, clearAllData
    };
  }
});

app.mount('#app');