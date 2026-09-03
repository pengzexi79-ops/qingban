const { createApp, nextTick } = Vue;

const ICON_PATHS = {
  message: '<path d="M5 6.5A3.5 3.5 0 0 1 8.5 3h7A3.5 3.5 0 0 1 19 6.5v4A3.5 3.5 0 0 1 15.5 14H12l-4.5 3v-3.17A3.5 3.5 0 0 1 5 10.5z"/><path d="M9 8.5h6M9 11h3"/>',
  users: '<path d="M16 19v-1.5a3.5 3.5 0 0 0-3.5-3.5h-5A3.5 3.5 0 0 0 4 17.5V19"/><circle cx="10" cy="7.5" r="3.5"/><path d="M16 7.5a3 3 0 0 1 0 5.8M18.5 14.5a3.5 3.5 0 0 1 2.5 3.3V19"/>',
  sparkle: '<path d="m12 3 1.25 4.3L17.5 9l-4.25 1.7L12 15l-1.25-4.3L6.5 9l4.25-1.7z"/><path d="m19 14 .6 2.1L21.5 17l-1.9.9L19 20l-.6-2.1-1.9-.9 1.9-.9zM5 15l.45 1.55L7 17l-1.55.45L5 19l-.45-1.55L3 17l1.55-.45z"/>',
  brain: '<path d="M9.5 4.5A3.5 3.5 0 0 0 6 8v.4a3.4 3.4 0 0 0-1.5 6.1A3.5 3.5 0 0 0 8 18h1.5"/><path d="M14.5 4.5A3.5 3.5 0 0 1 18 8v.4a3.4 3.4 0 0 1 1.5 6.1A3.5 3.5 0 0 1 16 18h-1.5M9.5 4.5v13M14.5 4.5v13M9.5 9h5M9.5 13h5"/>',
  settings: '<path d="M12 3.5v2M12 18.5v2M4.5 12h-2M21.5 12h-2M6 6 4.6 4.6M19.4 19.4 18 18M18 6l1.4-1.4M4.6 19.4 6 18"/><circle cx="12" cy="12" r="4.5"/>',
  palette: '<path d="M12 3.5a8.5 8.5 0 0 0 0 17h1.4a1.8 1.8 0 0 0 .8-3.4 1.8 1.8 0 0 1 .8-3.4H17a3.5 3.5 0 0 0 3.5-3.5A6.7 6.7 0 0 0 12 3.5Z"/><circle cx="7.5" cy="10" r=".8"/><circle cx="9.5" cy="6.8" r=".8"/><circle cx="14" cy="6.5" r=".8"/><circle cx="17" cy="9" r=".8"/>',
  search: '<circle cx="10.8" cy="10.8" r="6.3"/><path d="m16 16 4.2 4.2"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  edit: '<path d="m4 16.7-.7 3.3 3.3-.7L18.4 7.5a2.2 2.2 0 0 0-3.1-3.1z"/><path d="m14.2 5.8 3.1 3.1"/>',
  close: '<path d="m6 6 12 12M18 6 6 18"/>',
  check: '<path d="m5 12 4.2 4.2L19 6.5"/>',
  'arrow-left': '<path d="M19 12H5M11 18l-6-6 6-6"/>',
  'arrow-right': '<path d="M5 12h14M13 6l6 6-6 6"/>',
  send: '<path d="m21 3-7.2 18-3.8-7L3 10.2z"/><path d="M10 14 21 3"/>',
  heart: '<path d="M20.8 8.9c0 5.2-8.8 10.2-8.8 10.2S3.2 14.1 3.2 8.9A4.5 4.5 0 0 1 12 6.7a4.5 4.5 0 0 1 8.8 2.2Z"/>',
  clock: '<circle cx="12" cy="12" r="8.5"/><path d="M12 7v5l3.5 2"/>',
  trash: '<path d="M4.5 7h15M9 7V4.5h6V7M7 7l.8 12.5h8.4L17 7M10 10.5v6M14 10.5v6"/>',
  bell: '<path d="M18 9.5a6 6 0 0 0-12 0c0 7-2.5 7-2.5 8.5h17C20.5 16.5 18 16.5 18 9.5ZM10 21h4"/>',
  database: '<ellipse cx="12" cy="5.5" rx="7.5" ry="3"/><path d="M4.5 5.5v6c0 1.7 3.4 3 7.5 3s7.5-1.3 7.5-3v-6M4.5 11.5v6c0 1.7 3.4 3 7.5 3s7.5-1.3 7.5-3v-6"/>',
  download: '<path d="M12 4v10M8 10l4 4 4-4M5 19.5h14"/>',
  refresh: '<path d="M19 8.5A7.5 7.5 0 1 0 20 14"/><path d="M19 4v4.5h-4.5"/>',
  shield: '<path d="M12 3.5 19 6v5.2c0 4.2-2.8 7.4-7 9.3-4.2-1.9-7-5.1-7-9.3V6z"/><path d="m8.5 12 2.2 2.2 4.8-4.8"/>',
  info: '<circle cx="12" cy="12" r="9"/><path d="M12 10v6M12 7.2v.1"/>',
  paperclip: '<path d="m20 11.5-8.2 8.2a5 5 0 0 1-7.1-7.1l8.6-8.6a3.5 3.5 0 0 1 5 5L9.7 17.6a2 2 0 0 1-2.8-2.8l8-8"/>'
};

function icon(name) {
  return '<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">' + (ICON_PATHS[name] || ICON_PATHS.sparkle) + '</g></svg>';
}

function todayLabel() {
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
}

const ChatView = {
  props: {
    companion: { type: Object, required: true },
    messages: { type: Array, default: () => [] },
    isReplying: Boolean,
    inputMessage: { type: String, default: '' }
  },
  emits: ['update-input', 'send', 'quick', 'proactive', 'back', 'open-settings'],
  methods: {
    icon,
    avatarStyle(item) { return { background: item && item.color ? item.color : '#8d83ee' }; },
    formatMessageTime(timestamp) { return new Date(timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }); },
    quick(text) { this.$emit('quick', text); },
    send() { this.$emit('send'); },
    update(value) { this.$emit('update-input', value); }
  },
  template: `
    <div class="chat-view">
      <div class="chat-contact-bar">
        <div class="chat-contact-info">
          <button class="chat-back-mobile" @click="$emit('back')" aria-label="返回"><span v-html="icon('arrow-left')"></span></button>
          <div class="avatar" :style="avatarStyle(companion)">{{ companion.initial || companion.name.slice(0, 1) }}</div>
          <div><strong>{{ companion.name }}</strong><span><i class="status-dot"></i>{{ companion.online ? '正在青伴里' : '暂时安静' }}</span></div>
        </div>
        <div class="chat-contact-actions"><button class="ghost-action" @click="$emit('proactive')"><span v-html="icon('send')"></span><span class="desktop-only">让 TA 主动说句话</span><span class="mobile-only">主动消息</span></button><button class="round-action" @click="$emit('open-settings')" aria-label="聊天设置"><span v-html="icon('settings')"></span></button></div>
      </div>
      <div class="chat-memory-hint"><span v-html="icon('brain')"></span><span>已为你参考 {{ companion.memories.length }} 条长期记忆</span><button @click="$emit('open-settings')">管理记忆 <span v-html="icon('arrow-right')"></span></button></div>
      <div class="chat-messages" ref="messageContainer">
        <div v-if="messages.length === 0" class="chat-empty"><div class="empty-orb"><span v-html="icon('heart')"></span></div><strong>从一句“嗨”开始吧</strong><p>{{ companion.greeting }}</p><button class="quick-chip" @click="quick('嗨，今天想和你聊聊')">发送问候</button></div>
        <template v-else>
          <div v-for="(message, index) in messages" :key="message.id" class="message-block" :class="{ mine: message.role === 'user' }">
            <div v-if="index === 0 || new Date(message.timestamp).toDateString() !== new Date(messages[index - 1].timestamp).toDateString()" class="date-divider">{{ new Date(message.timestamp).toLocaleDateString('zh-CN', { month: 'long', day: 'numeric' }) }}</div>
            <div class="message-line">
              <div v-if="message.role !== 'user'" class="message-avatar avatar" :style="avatarStyle(companion)">{{ companion.initial }}</div>
              <div class="message-body"><div class="message-bubble">{{ message.content }}</div><time>{{ formatMessageTime(message.timestamp) }}<span v-if="message.proactive" class="proactive-tag">主动</span></time></div>
              <div v-if="message.role === 'user'" class="message-avatar user-avatar">我</div>
            </div>
          </div>
          <div v-if="isReplying" class="message-block"><div class="message-line"><div class="message-avatar avatar" :style="avatarStyle(companion)">{{ companion.initial }}</div><div class="message-body"><div class="message-bubble typing"><i></i><i></i><i></i></div></div></div></div>
        </template>
      </div>
      <div class="chat-quick-replies" v-if="messages.length < 4 && !isReplying"><button class="quick-chip" @click="quick('今天有点累')">今天有点累</button><button class="quick-chip" @click="quick('我想分享一件小事')">分享一件小事</button><button class="quick-chip" @click="quick('你还记得我吗？')">你还记得我吗？</button></div>
      <div class="chat-composer"><button class="composer-icon" aria-label="添加附件"><span v-html="icon('paperclip')"></span></button><div class="composer-input"><textarea :value="inputMessage" @input="update($event.target.value)" @keydown.enter.exact.prevent="send" rows="1" placeholder="说点什么，青伴在听…"></textarea><span>{{ inputMessage.length }}/500</span></div><button class="send-button" :class="{ ready: inputMessage.trim() }" :disabled="!inputMessage.trim() || isReplying" @click="send" aria-label="发送消息"><span v-html="icon('send')"></span></button></div>
      <p class="chat-disclaimer"><span v-html="icon('shield')"></span> 当前为前端演示，对话回复由本地模拟生成</p>
    </div>
  `
};

const app = createApp({
  components: { ChatView },
  data() {
    const state = QingbanStore.getState();
    return {
      activeView: 'inbox', activeChatId: null, inboxFilter: 'all', searchQuery: '', memoryQuery: '', memoryFilter: 'all',
      companions: state.companions, settings: state.settings, inputMessage: '', isReplying: false,
      showCompanionDialog: false, showMemoryDialog: false, showProfileDialog: false, showAbout: false,
      editingCompanion: null, editingMemory: null, toastMessage: '', toastTimer: null, autoMessageTimer: null,
      companionForm: thisBlankCompanion(), memoryForm: thisBlankMemory(state.companions),
      profileForm: { nickname: state.settings.nickname, signature: state.settings.signature },
      avatarColors: ['#8d83ee', '#ef9d98', '#66b7ae', '#e4ae6b', '#7296d9', '#ad87bd'],
      navItems: [
        { id: 'inbox', label: '会话', icon: 'message' }, { id: 'friends', label: '通讯录', icon: 'users' },
        { id: 'companions', label: 'AI 好友', icon: 'sparkle' }, { id: 'memories', label: '记忆库', icon: 'brain' }, { id: 'settings', label: '设置', icon: 'settings' }
      ],
      memoryFilters: [{ id: 'all', label: '全部' }, { id: 'preference', label: '偏好' }, { id: 'event', label: '事件' }, { id: 'relationship', label: '关系' }]
    };
  },
  computed: {
    mobileNavItems() { return this.navItems.filter(item => ['inbox', 'friends', 'companions', 'settings'].includes(item.id)); },
    activeCompanion() { return this.companions.find(item => item.id === this.activeChatId) || null; },
    activeMessages() { return this.activeCompanion ? (this.activeCompanion.messages || []) : []; },
    unreadTotal() { return this.companions.reduce((sum, item) => sum + (item.unread ? 1 : 0), 0); },
    memoryCount() { return this.companions.reduce((sum, item) => sum + (item.memories || []).length, 0); },
    proactiveCount() { return this.companions.filter(item => item.canAutoMessage).length; },
    userInitial() { return (this.settings.nickname || '你').slice(0, 1); },
    greetingText() {
      const hour = new Date().getHours();
      if (hour < 6) return '夜深了，也要记得照顾自己';
      if (hour < 12) return '早上好，愿今天从轻松开始';
      if (hour < 18) return '下午好，给自己留一点呼吸';
      return '晚上好，今天也辛苦了';
    },
    viewTitle() {
      if (this.activeView === 'inbox') return '会话';
      if (this.activeView === 'friends') return '通讯录';
      if (this.activeView === 'companions') return 'AI 好友';
      if (this.activeView === 'memories') return '记忆库';
      if (this.activeView === 'settings') return '我的青伴';
      return this.activeCompanion ? this.activeCompanion.name : '青伴';
    },
    filteredCompanions() {
      const query = this.searchQuery.trim().toLowerCase();
      return this.companions.filter(item => {
        const matchFilter = this.inboxFilter !== 'unread' || item.unread;
        const last = this.lastMessagePreview(item.id).toLowerCase();
        return matchFilter && (!query || item.name.toLowerCase().includes(query) || last.includes(query));
      }).sort((a, b) => this.lastTimestamp(b.id) - this.lastTimestamp(a.id));
    },
    allMemories() { return this.companions.flatMap(companion => (companion.memories || []).map(memory => ({ ...memory, companion }))); },
    filteredMemories() {
      const query = this.memoryQuery.trim().toLowerCase();
      return this.allMemories.filter(memory => {
        const typeMatch = this.memoryFilter === 'all' || memory.type === this.memoryFilter;
        const text = [memory.title, memory.content, memory.companion.name, memory.source].join(' ').toLowerCase();
        return typeMatch && (!query || text.includes(query));
      }).sort((a, b) => this.memoryDateValue(b.date) - this.memoryDateValue(a.date));
    }
  },
  watch: {
    settings: { deep: true, handler(value) { QingbanStore.saveSettings(value); } },
    activeMessages() { this.$nextTick(() => this.scrollChatToBottom()); }
  },
  mounted() { this.$nextTick(() => this.scrollChatToBottom()); this.startAutoMessages(); },
  beforeUnmount() { if (this.autoMessageTimer) window.clearInterval(this.autoMessageTimer); },
  methods: {
    icon,
    navigate(view) { this.activeView = view; if (view !== 'inbox') this.activeChatId = null; this.searchQuery = ''; this.$nextTick(() => this.scrollChatToBottom()); },
    focusSearch() { this.$nextTick(() => { const input = this.$refs.searchField && this.$refs.searchField.querySelector('input'); if (input) input.focus(); }); },
    openChat(companion) {
      if (!companion) return;
      this.activeView = 'inbox'; this.activeChatId = companion.id; this.inputMessage = '';
      QingbanStore.markRead(companion.id); companion.unread = false;
      if (!companion.messages || companion.messages.length === 0) { QingbanStore.addMessage(companion.id, { role: 'assistant', content: companion.greeting }); this.refreshState(); }
      this.$nextTick(() => this.scrollChatToBottom());
    },
    closeChat() { this.activeChatId = null; this.inputMessage = ''; },
    scrollChatToBottom() { document.querySelectorAll('.chat-messages').forEach(container => { container.scrollTop = container.scrollHeight; }); },
    refreshState() { const state = QingbanStore.getState(); this.companions = state.companions; this.settings = state.settings; this.profileForm = { nickname: state.settings.nickname, signature: state.settings.signature }; },
    avatarStyle(item) { return { background: item && item.color ? item.color : '#8d83ee' }; },
    lastTimestamp(id) { const companion = this.companions.find(item => item.id === id); return companion && companion.messages && companion.messages.length ? companion.messages[companion.messages.length - 1].timestamp : 0; },
    lastMessagePreview(id) { const companion = this.companions.find(item => item.id === id); if (!companion || !companion.messages || !companion.messages.length) return '还没有聊天记录，去打个招呼吧'; const message = companion.messages[companion.messages.length - 1]; return (message.role === 'user' ? '你：' : '') + message.content; },
    lastMessageTime(id) { const timestamp = this.lastTimestamp(id); if (!timestamp) return ''; const date = new Date(timestamp); const now = new Date(); if (date.toDateString() === now.toDateString()) return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }); return date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' }); },
    openNewCompanion() { this.editingCompanion = null; this.companionForm = thisBlankCompanion(); this.showCompanionDialog = true; },
    openCompanionEditor(companion) {
      if (!companion) return;
      this.editingCompanion = companion;
      this.companionForm = { id: companion.id, name: companion.name, initial: companion.initial, color: companion.color, category: companion.category, tagline: companion.tagline, personality: companion.personality, greeting: companion.greeting, canAutoMessage: companion.canAutoMessage, frequency: companion.frequency, online: companion.online, unread: companion.unread, memories: companion.memories, messages: companion.messages };
      this.showCompanionDialog = true;
    },
    saveCompanion() {
      const form = { ...this.companionForm };
      if (!form.name.trim()) return this.showToast('先给 AI 好友起个名字吧');
      if (!form.tagline.trim()) form.tagline = form.category + '，随时可以来聊聊';
      if (!form.greeting.trim()) form.greeting = '嗨，我已经准备好听你说今天的事了。';
      form.initial = form.name.slice(0, 1); form.frequency = form.canAutoMessage ? '偶尔想起你' : '已暂停'; form.online = true;
      let item;
      if (this.editingCompanion) { item = QingbanStore.upsertCompanion(form); this.showToast('AI 好友资料已更新'); }
      else { item = QingbanStore.upsertCompanion({ ...form, id: null, memories: [], messages: [], unread: false }); QingbanStore.addMessage(item.id, { role: 'assistant', content: form.greeting }); this.showToast('AI 好友已创建'); }
      this.refreshState(); this.showCompanionDialog = false; this.editingCompanion = null; this.activeChatId = item.id; this.activeView = 'inbox'; this.$nextTick(() => this.scrollChatToBottom());
    },
    removeCompanion(companion) {
      if (!companion || !window.confirm('确定删除“' + companion.name + '”吗？聊天记录和长期记忆也会一起删除。')) return;
      QingbanStore.removeCompanion(companion.id); if (this.activeChatId === companion.id) this.activeChatId = null; this.refreshState(); this.showToast('已删除 ' + companion.name);
    },
    sendMessage() {
      const text = this.inputMessage.trim(); if (!text || !this.activeCompanion || this.isReplying) return;
      const companionId = this.activeCompanion.id; QingbanStore.addMessage(companionId, { role: 'user', content: text }); this.inputMessage = ''; this.refreshState(); this.isReplying = true; this.$nextTick(() => this.scrollChatToBottom()); this.maybeCaptureMemory(text, companionId);
      window.setTimeout(() => { const current = this.companions.find(item => item.id === companionId); if (!current) return; QingbanStore.addMessage(companionId, { role: 'assistant', content: this.generateReply(text, current) }); this.isReplying = false; this.refreshState(); this.$nextTick(() => this.scrollChatToBottom()); }, 650 + Math.round(Math.random() * 500));
    },
    sendQuickMessage(text) { this.inputMessage = text; this.sendMessage(); },
    generateReply(text) {
      const value = text.toLowerCase();
      if (/累|疲惫|难过|不开心|焦虑|压力/.test(value)) return '我听见了，你现在好像真的有点辛苦。先不用急着把一切处理好，跟我说说最让你累的那一小部分，好吗？';
      if (/记得|忘了|还记得/.test(value)) return '记得呀。你愿意告诉我的事情，我会认真放在属于我们的记忆里；如果哪天不想保留，也可以随时删掉。';
      if (/喜欢|爱吃|偏好|习惯/.test(value)) return '好，我把这个小偏好记下来了。以后聊到相关的事，我会尽量用你舒服的方式陪你。';
      if (/早安|早上/.test(value)) return '早安。今天不用一下子变得很厉害，先把第一件小事做完就很好。';
      if (/晚安|睡觉/.test(value)) return '晚安，今天发生的一切先放在这里。去休息吧，明天回来时，我还会接着听。';
      if (/谢谢|感谢/.test(value)) return '不用客气，能陪你把这句话说完，对我来说就很重要。';
      const replies = ['嗯，我在听。你愿意从哪里开始都可以。', '这件事对你来说一定有特别的分量。要不要再和我说一点？', '我先不急着给结论，想和你一起把感受理清楚。', '收到，已经放进我们今天的对话里了。你现在感觉好一点了吗？', '如果你愿意，我可以一直陪你把这件事慢慢说完。'];
      return replies[Math.floor(Math.random() * replies.length)];
    },
    maybeCaptureMemory(text, companionId) {
      if (!/记住|喜欢|习惯|生日|重要|下周|明天|讨厌/.test(text) || text.length < 5) return;
      const companion = this.companions.find(item => item.id === companionId); if (!companion) return;
      if ((companion.memories || []).some(memory => memory.content === text)) return;
      QingbanStore.upsertMemory(companionId, { type: /生日|下周|明天|重要/.test(text) ? 'event' : 'preference', title: '从对话中发现的小事', content: text, date: todayLabel(), source: '对话整理' }); this.refreshState(); this.showToast('已把这件小事放进记忆库');
    },
    sendProactiveMessage(companion) {
      if (!companion) return;
      const pools = { '温柔陪伴': ['刚刚路过一阵风，突然想问问你今天还好吗？', '我给你留了一盏小灯。忙完的时候，记得回来坐一会儿。'], '知心朋友': ['突然想起你了，今天有没有发生什么值得分享的小事？', '不打扰你，只是想说：今天也辛苦了。'], '元气搭子': ['叮！你的今日份陪伴已送达，准备好说点什么了吗？', '来报到一下！今天的心情是晴天、阴天，还是想下点小雨？'], '安静倾听': ['我在这里，不用组织好语言，想到什么就发给我。'], '灵感导师': ['刚刚有个问题想和你一起想：今天什么事最值得被留下？'] };
      const list = pools[companion.category] || pools['温柔陪伴']; QingbanStore.addMessage(companion.id, { role: 'assistant', content: list[Math.floor(Math.random() * list.length)], proactive: true }); this.refreshState(); if (this.activeChatId === companion.id) this.$nextTick(() => this.scrollChatToBottom()); this.showToast(companion.name + ' 主动来找你了');
    },
    startAutoMessages() {
      this.autoMessageTimer = window.setInterval(() => {
        if (!this.settings.autoMessages || !this.companions.length || document.hidden) return;
        const hour = new Date().getHours(); if (this.settings.quietHours && (hour >= 23 || hour < 8)) return;
        const eligible = this.companions.filter(item => item.canAutoMessage); if (!eligible.length || Math.random() > 0.45) return;
        this.sendProactiveMessage(eligible[Math.floor(Math.random() * eligible.length)]);
      }, 45000);
    },
    openNewMemory() { if (!this.companions.length) return this.showToast('请先创建一个 AI 好友'); this.editingMemory = null; this.memoryForm = thisBlankMemory(this.companions); this.showMemoryDialog = true; },
    openMemoryEditor(memory) { this.editingMemory = memory; this.memoryForm = { id: memory.id, companionId: memory.companion.id, type: memory.type, title: memory.title, content: memory.content, date: memory.date, source: memory.source }; this.showMemoryDialog = true; },
    saveMemory() { if (!this.memoryForm.companionId || !this.memoryForm.title.trim() || !this.memoryForm.content.trim()) return this.showToast('把记忆的标题和内容补充完整吧'); QingbanStore.upsertMemory(this.memoryForm.companionId, { ...this.memoryForm, source: this.memoryForm.source || '手动添加' }); this.refreshState(); this.showMemoryDialog = false; this.editingMemory = null; this.showToast('记忆库已更新'); },
    removeMemory(memory) { if (!memory || !window.confirm('确定删除这条记忆吗？删除后，AI 将不再参考它。')) return; QingbanStore.removeMemory(memory.companion.id, memory.id); this.refreshState(); this.showToast('记忆已删除'); },
    memoryTypeLabel(type) { return ({ preference: '偏好', event: '事件', relationship: '关系' })[type] || '记忆'; },
    memoryIcon(type) { return ({ preference: 'heart', event: 'clock', relationship: 'users' })[type] || 'brain'; },
    memoryDateValue(date) { return new Date(String(date).replace(/年|月/g, '-').replace(/日/g, '')).getTime() || 0; },
    toggleSetting(key) { this.settings[key] = !this.settings[key]; this.showToast(this.settings[key] ? '已开启' : '已关闭'); },
    editProfile() { this.profileForm = { nickname: this.settings.nickname, signature: this.settings.signature }; this.showProfileDialog = true; },
    saveProfile() { this.settings.nickname = this.profileForm.nickname.trim() || '林林'; this.settings.signature = this.profileForm.signature.trim() || '把日子过成被记住的样子'; this.showProfileDialog = false; this.showToast('个人资料已更新'); },
    exportData() { const blob = new Blob([JSON.stringify(QingbanStore.exportData(), null, 2)], { type: 'application/json;charset=utf-8' }); const url = URL.createObjectURL(blob); const link = document.createElement('a'); link.href = url; link.download = 'qingban-data-' + new Date().toISOString().slice(0, 10) + '.json'; link.click(); URL.revokeObjectURL(url); this.showToast('数据备份已导出'); },
    resetDemo() { if (!window.confirm('恢复演示数据会覆盖当前本机数据，确定继续吗？')) return; const state = QingbanStore.reset(); this.companions = state.companions; this.settings = state.settings; this.activeChatId = null; this.showToast('演示数据已恢复'); },
    clearAllData() { if (!window.confirm('确定清空当前浏览器中的所有青伴数据吗？此操作不可恢复。')) return; QingbanStore.saveState({ companions: [], settings: { ...this.settings } }); this.companions = []; this.activeChatId = null; this.showToast('本机数据已清空'); },
    closeModal() { this.showCompanionDialog = false; this.showMemoryDialog = false; this.showProfileDialog = false; this.editingCompanion = null; this.editingMemory = null; },
    showToast(message) { this.toastMessage = message; if (this.toastTimer) window.clearTimeout(this.toastTimer); this.toastTimer = window.setTimeout(() => { this.toastMessage = ''; }, 2400); }
  }
});

function thisBlankCompanion() { return { name: '', initial: '', color: '#8d83ee', category: '温柔陪伴', tagline: '', personality: '', greeting: '', canAutoMessage: true, memoryMode: 'curated' }; }
function thisBlankMemory(companions) { return { id: null, companionId: companions && companions.length ? companions[0].id : '', type: 'preference', title: '', content: '', date: todayLabel(), source: '手动添加' }; }

app.mount('#app');
