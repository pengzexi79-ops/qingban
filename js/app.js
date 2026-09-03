/**
 * 青伴 - 主应用逻辑 (Vue 3)
 */

const { createApp, ref, computed, watch, onMounted, nextTick } = Vue;

const app = createApp({
  setup() {
    // === 页面状态 ===
    const currentPage = ref('chatlist');
    const searchQuery = ref('');
    const showChatMenu = ref(false);

    // === 数据 ===
    const companions = ref(Store.getCompanions());
    const currentChatId = ref(null);
    const inputMessage = ref('');
    const messageList = ref(null);
    const editingCompanion = ref(null);
    const userSettings = ref(Store.getSettings());
    const pageHistory = ref([]);

    // 头像颜色选项
    const avatarColors = [
      '#FF6B6B', '#FF8E53', '#FFC857', '#57C47A',
      '#4ECDC4', '#45B7D1', '#6C5CE7', '#A855F7',
      '#EC4899', '#8B5CF6', '#06B6D4', '#10B981'
    ];

    // 创建伴侣表单
    const companionForm = ref(getEmptyForm());

    function getEmptyForm() {
      return {
        name: '',
        avatarColor: '#57C47A',
        personality: '',
        roleSetting: '',
        greeting: '',
        canAutoMessage: true
      };
    }

    // === 计算属性 ===
    const pageTitle = computed(() => {
      const titles = {
        chatlist: '青伴',
        companions: 'AI伴侣',
        settings: '设置'
      };
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
      return currentCompanion.value?.name || '';
    });

    const currentMessages = computed(() => {
      if (!currentChatId.value) return [];
      return Store.getChatMessages(currentChatId.value);
    });

    // === 导航方法 ===
    function goAddCompanion() {
      editingCompanion.value = null;
      companionForm.value = getEmptyForm();
      pageHistory.value.push(currentPage.value);
      currentPage.value = 'addCompanion';
    }

    function goBack() {
      const prev = pageHistory.value.pop();
      if (prev) {
        currentPage.value = prev;
      } else {
        currentPage.value = 'chatlist';
      }
    }

    // === 聊天方法 ===
    function openChat(companion) {
      currentChatId.value = companion.id;
      pageHistory.value.push(currentPage.value);
      currentPage.value = 'chat';

      // 如果没有消息，发送开场白
      const messages = Store.getChatMessages(companion.id);
      if (messages.length === 0 && companion.greeting) {
        Store.addMessage(companion.id, {
          role: 'assistant',
          content: companion.greeting
        });
      }

      nextTick(() => {
        scrollToBottom();
      });
    }

    function sendMessage() {
      const text = inputMessage.value.trim();
      if (!text || !currentChatId.value) return;

      // 添加用户消息
      Store.addMessage(currentChatId.value, {
        role: 'user',
        content: text
      });
      inputMessage.value = '';

      // 模拟AI回复（初版框架，后续接入真实AI）
      setTimeout(() => {
        const reply = generateReply(text, currentCompanion.value);
        Store.addMessage(currentChatId.value, {
          role: 'assistant',
          content: reply
        });
        // 触发响应式更新
        companions.value = [...companions.value];
        nextTick(() => {
          scrollToBottom();
        });
      }, 500 + Math.random() * 1000);
    }

    function scrollToBottom() {
      if (messageList.value) {
        messageList.value.scrollTop = messageList.value.scrollHeight;
      }
    }

    // === 伴侣管理 ===
    function saveCompanion() {
      const form = companionForm.value;
      if (!form.name.trim()) {
        alert('请输入伴侣名字');
        return;
      }

      if (editingCompanion.value) {
        // 更新现有伴侣
        Store.updateCompanion(editingCompanion.value.id, {
          name: form.name,
          avatarColor: form.avatarColor,
          personality: form.personality,
          roleSetting: form.roleSetting,
          greeting: form.greeting,
          canAutoMessage: form.canAutoMessage
        });
      } else {
        // 创建新伴侣
        const newCompanion = Store.addCompanion({
          name: form.name,
          avatarColor: form.avatarColor,
          personality: form.personality,
          roleSetting: form.roleSetting,
          greeting: form.greeting,
          canAutoMessage: form.canAutoMessage
        });

        // 添加开场白
        if (form.greeting) {
          Store.addMessage(newCompanion.id, {
            role: 'assistant',
            content: form.greeting
          });
        }
      }

      companions.value = Store.getCompanions();
      goBack();
    }

    function editCompanion(companion) {
      editingCompanion.value = companion;
      companionForm.value = {
        name: companion.name,
        avatarColor: companion.avatarColor,
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

    // === 消息辅助 ===
    function getLastMessage(companionId) {
      const messages = Store.getChatMessages(companionId);
      if (messages.length === 0) return '暂无消息';
      const last = messages[messages.length - 1];
      const prefix = last.role === 'user' ? '[我] ' : '';
      return prefix + last.content;
    }

    function getLastTime(companionId) {
      const messages = Store.getChatMessages(companionId);
      if (messages.length === 0) return '';
      return formatTime(messages[messages.length - 1].timestamp);
    }

    function getUnreadCount(companionId) {
      // 初版暂不支持未读计数
      return 0;
    }

    function formatTime(timestamp) {
      const date = new Date(timestamp);
      const now = new Date();
      const isToday = date.toDateString() === now.toDateString();

      if (isToday) {
        return date.getHours().toString().padStart(2, '0') + ':' +
               date.getMinutes().toString().padStart(2, '0');
      }

      const yesterday = new Date(now);
      yesterday.setDate(yesterday.getDate() - 1);
      if (date.toDateString() === yesterday.toDateString()) {
        return '昨天';
      }

      return (date.getMonth() + 1) + '/' + date.getDate();
    }

    // === 设置方法 ===
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

    // === AI回复生成（初版模拟） ===
    function generateReply(userMsg, companion) {
      if (!companion) return '...';

      const personality = companion.personality || '';
      const replies = {
        default: [
          '我在听呢，继续说~',
          '嗯嗯，我理解你的感受。',
          '谢谢你愿意和我分享这些。',
          '这确实是很正常的感受呢。',
          '你今天过得怎么样？',
          '有什么我可以帮到你的吗？',
          '我一直在这里陪着你哦。',
          '听起来你经历了很多呢。'
        ],
        warm: [
          '抱抱你~',
          '你真的很棒！',
          '我相信你可以的！',
          '无论发生什么，我都会在你身边。',
          '你对我来说很重要。',
          '想你了，最近还好吗？'
        ],
        calm: [
          '我理解。',
          '这确实值得思考。',
          '你的想法很有道理。',
          '让我们一起想想办法。',
          '慢慢来，不急。',
          '有时候放慢脚步也是种智慧。'
        ]
      };

      // 根据性格关键词选择回复风格
      let pool = replies.default;
      if (personality.includes('温柔') || personality.includes('甜') || personality.includes('暖')) {
        pool = replies.warm;
      } else if (personality.includes('冷静') || personality.includes('理性') || personality.includes('沉稳')) {
        pool = replies.calm;
      }

      // 特定关键词回复
      if (userMsg.includes('你好') || userMsg.includes('hi') || userMsg.includes('hello')) {
        return '你好呀！很高兴见到你~';
      }
      if (userMsg.includes('再见') || userMsg.includes('bye')) {
        return '下次见哦，我会想你的~';
      }
      if (userMsg.includes('喜欢') || userMsg.includes('爱')) {
        return '我也很喜欢和你聊天呢~';
      }
      if (userMsg.includes('难过') || userMsg.includes('伤心') || userMsg.includes('不开心')) {
        return '抱抱你，会好起来的。我在这里陪你。';
      }
      if (userMsg.includes('开心') || userMsg.includes('高兴') || userMsg.includes('快乐')) {
        return '看到你开心，我也很高兴呢！';
      }
      if (userMsg.includes('?') || userMsg.includes('？')) {
        return '这是个好问题，让我想想...';
      }

      return pool[Math.floor(Math.random() * pool.length)];
    }

    // === 主动消息模拟 ===
    let autoMessageTimer = null;

    function startAutoMessages() {
      // 每 30-120 秒随机发送一条主动消息
      const scheduleNext = () => {
        const delay = 30000 + Math.random() * 90000; // 30-120秒
        autoMessageTimer = setTimeout(() => {
          sendAutoMessage();
          scheduleNext();
        }, delay);
      };
      scheduleNext();
    }

    function sendAutoMessage() {
      if (!userSettings.value.autoMsgNotify) return;

      const eligible = companions.value.filter(c => c.canAutoMessage !== false);
      if (eligible.length === 0) return;

      const companion = eligible[Math.floor(Math.random() * eligible.length)];
      const autoMessages = [
        '在吗？想你了~',
        '今天过得怎么样？',
        '有空聊聊天吗？',
        '刚刚想到你了~',
        '天气不错呢，心情也好~',
        '你在忙什么呢？',
        '好久没聊了，想和你说说话。'
      ];

      const msg = autoMessages[Math.floor(Math.random() * autoMessages.length)];
      Store.addMessage(companion.id, {
        role: 'assistant',
        content: msg
      });

      // 更新列表以显示最新消息
      companions.value = [...companions.value];
    }

    // === 监听设置变化并保存 ===
    watch(userSettings, (newVal) => {
      Store.saveSettings(newVal);
    }, { deep: true });

    // === 生命周期 ===
    onMounted(() => {
      // 启动主动消息功能
      startAutoMessages();

      // 如果有伴侣但没有开场消息，补充开场白
      companions.value.forEach(c => {
        const msgs = Store.getChatMessages(c.id);
        if (msgs.length === 0 && c.greeting) {
          Store.addMessage(c.id, {
            role: 'assistant',
            content: c.greeting
          });
        }
      });
    });

    return {
      // 状态
      currentPage,
      searchQuery,
      showChatMenu,
      companions,
      currentChatId,
      inputMessage,
      messageList,
      editingCompanion,
      userSettings,
      companionForm,
      avatarColors,

      // 计算属性
      pageTitle,
      filteredCompanions,
      currentCompanion,
      currentChatName,
      currentMessages,

      // 方法
      goAddCompanion,
      goBack,
      openChat,
      sendMessage,
      saveCompanion,
      editCompanion,
      deleteCompanion,
      getLastMessage,
      getLastTime,
      getUnreadCount,
      formatTime,
      exportMemory,
      clearAllData
    };
  }
});

app.mount('#app');
