/**
 * 青伴 - 前端演示数据层
 * 只负责 localStorage，不包含任何真实 AI / API 调用。
 */
(function () {
  const KEY = 'qingban_frontend_v2';
  const DAY = 24 * 60 * 60 * 1000;

  const seedCompanions = [
    {
      id: 'mumu',
      name: '沐沐',
      initial: '沐',
      color: '#8d83ee',
      category: '温柔陪伴',
      tagline: '会记得你小习惯的温柔朋友',
      personality: '温柔、耐心，不急着给答案。先听我把话说完，偶尔用轻松的方式逗我开心。',
      greeting: '晚上好呀，我已经把一盏小灯留给你了。今天过得怎么样？',
      canAutoMessage: true,
      frequency: '偶尔想起你',
      online: true,
      unread: true,
      memories: [
        { id: 'mumu-1', type: 'preference', title: '喜欢雨天听歌', content: '下雨的时候，喜欢戴耳机听舒缓的歌，不太想被催着做决定。', date: '2026年09月02日', source: '你告诉我的' },
        { id: 'mumu-2', type: 'event', title: '最近在准备一个重要计划', content: '最近有一件重要的事在准备中，希望我多给你一点鼓励，少一些说教。', date: '2026年09月01日', source: '对话整理' }
      ],
      messages: [
        { id: 'mumu-msg-1', role: 'assistant', content: '晚上好呀，我已经把一盏小灯留给你了。今天过得怎么样？', timestamp: Date.now() - 2 * 60 * 60 * 1000 },
        { id: 'mumu-msg-2', role: 'user', content: '今天有点累，但还是把事情做完了。', timestamp: Date.now() - 90 * 60 * 1000 },
        { id: 'mumu-msg-3', role: 'assistant', content: '那你今天已经很棒了。把“做完”也算作值得被看见的进度，剩下的我们明天再慢慢想。', timestamp: Date.now() - 88 * 60 * 1000 }
      ]
    },
    {
      id: 'xiaoman',
      name: '小满',
      initial: '满',
      color: '#ef9d98',
      category: '知心朋友',
      tagline: '在你需要时，安静地站在这一边',
      personality: '真诚直接，但不冒犯。会认真记住重要的事，也会提醒我照顾自己的感受。',
      greeting: '嗨，我是小满。今天想从哪一件小事开始聊？',
      canAutoMessage: true,
      frequency: '每天一次',
      online: true,
      unread: false,
      memories: [
        { id: 'xiaoman-1', type: 'event', title: '周末想去走走', content: '周末如果天气不错，想去附近公园散步，不安排太满。', date: '2026年08月31日', source: '你告诉我的' }
      ],
      messages: [
        { id: 'xiaoman-msg-1', role: 'assistant', content: '周末如果天气不错，一起去公园散散心？', timestamp: Date.now() - DAY },
        { id: 'xiaoman-msg-2', role: 'user', content: '好呀，先记下来。', timestamp: Date.now() - DAY + 10 * 60 * 1000 }
      ]
    },
    {
      id: 'xiaoyu',
      name: '小屿',
      initial: '屿',
      color: '#66b7ae',
      category: '灵感导师',
      tagline: '帮你把脑海里的想法慢慢理清',
      personality: '清晰、有好奇心，喜欢用问题陪我思考，不会把所有事情都变成任务。',
      greeting: '欢迎来到小屿的房间。今天有什么想法，值得被好好放在桌面上？',
      canAutoMessage: false,
      frequency: '已暂停',
      online: false,
      unread: false,
      memories: [],
      messages: []
    }
  ];

  const defaultSettings = {
    nickname: '林林',
    signature: '把日子过成被记住的样子',
    autoMessages: true,
    notifications: true,
    quietHours: true,
    dimMode: false
  };

  function clone(value) {
    return JSON.parse(JSON.stringify(value));
  }

  function read() {
    try {
      const raw = localStorage.getItem(KEY);
      if (!raw) return { companions: clone(seedCompanions), settings: clone(defaultSettings) };
      const parsed = JSON.parse(raw);
      return {
        companions: Array.isArray(parsed.companions) ? parsed.companions : clone(seedCompanions),
        settings: { ...defaultSettings, ...(parsed.settings || {}) }
      };
    } catch (error) {
      return { companions: clone(seedCompanions), settings: clone(defaultSettings) };
    }
  }

  function write(data) {
    localStorage.setItem(KEY, JSON.stringify(data));
  }

  function createId(prefix) {
    return prefix + '-' + Date.now().toString(36) + Math.random().toString(36).slice(2, 7);
  }

  window.QingbanStore = {
    key: KEY,
    seed() { return clone({ companions: seedCompanions, settings: defaultSettings }); },
    getState() { return read(); },
    saveState(data) { write(data); },
    reset() { const data = this.seed(); write(data); return data; },
    createId,
    clone,
    getCompanion(id) { return read().companions.find(item => item.id === id) || null; },
    upsertCompanion(companion) {
      const data = read();
      const item = { ...companion };
      if (!item.id) item.id = createId('companion');
      if (!Array.isArray(item.memories)) item.memories = [];
      if (!Array.isArray(item.messages)) item.messages = [];
      const index = data.companions.findIndex(entry => entry.id === item.id);
      if (index === -1) data.companions.push(item); else data.companions.splice(index, 1, item);
      write(data);
      return item;
    },
    removeCompanion(id) {
      const data = read();
      data.companions = data.companions.filter(item => item.id !== id);
      write(data);
    },
    addMessage(companionId, message) {
      const data = read();
      const companion = data.companions.find(item => item.id === companionId);
      if (!companion) return null;
      const entry = { id: createId('message'), timestamp: Date.now(), ...message };
      companion.messages = Array.isArray(companion.messages) ? companion.messages : [];
      companion.messages.push(entry);
      companion.unread = entry.role === 'assistant';
      write(data);
      return entry;
    },
    markRead(companionId) {
      const data = read();
      const companion = data.companions.find(item => item.id === companionId);
      if (companion) companion.unread = false;
      write(data);
    },
    upsertMemory(companionId, memory) {
      const data = read();
      const companion = data.companions.find(item => item.id === companionId);
      if (!companion) return null;
      const entry = { id: memory.id || createId('memory'), date: memory.date || new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date()), source: memory.source || '手动添加', ...memory };
      companion.memories = Array.isArray(companion.memories) ? companion.memories : [];
      const index = companion.memories.findIndex(item => item.id === entry.id);
      if (index === -1) companion.memories.unshift(entry); else companion.memories.splice(index, 1, entry);
      write(data);
      return entry;
    },
    removeMemory(companionId, memoryId) {
      const data = read();
      const companion = data.companions.find(item => item.id === companionId);
      if (companion) companion.memories = (companion.memories || []).filter(item => item.id !== memoryId);
      write(data);
    },
    saveSettings(settings) {
      const data = read();
      data.settings = { ...data.settings, ...settings };
      write(data);
      return data.settings;
    },
    exportData() { return { ...read(), exportedAt: new Date().toISOString(), version: 'qingban-frontend-v2' }; }
  };
})();
