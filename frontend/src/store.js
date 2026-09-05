/**
 * 亲伴 - 前端原型数据层
 * 只在浏览器 localStorage 中保存演示数据，不包含真实数据库、向量库或后台服务。
 */
const KEY = 'qinban_frontend_v4';
  const LEGACY_KEYS = ['qingban_frontend_v4', 'qingban_frontend_v3', 'qingban_frontend_v2'];
  const DAY = 24 * 60 * 60 * 1000;
  const now = Date.now();

  const defaultPersona = {
    identity: '一位愿意长期陪伴、尊重用户边界的 AI 好友',
    relationship: '知心朋友',
    personality: '温柔、真诚、耐心，不急着给答案。',
    speakingStyle: '自然口语，先回应感受，再给简短建议。',
    boundaries: '不替用户做高风险决定；遇到危机内容优先建议寻求现实帮助。',
    forbiddenTopics: '不诱导依赖，不冒充现实中的特定人物。'
  };

  const defaultMemorySettings = {
    enabled: true,
    mode: 'hybrid',
    summaryMode: 'rolling',
    timeRangeDays: 365,
    searchThreshold: 0.65,
    maxItems: 12
  };

  const defaultChatStyle = {
    markdown: true,
    streaming: true,
    typing: true,
    splitMessages: true,
    replyDelay: 650,
    bubbleStyle: 'soft'
  };

  const defaultProactive = {
    enabled: true,
    start: '09:00',
    end: '22:30',
    frequency: 'balanced',
    minMinutes: 45,
    maxMinutes: 240,
    dailyLimit: 4,
    avoidBusyTime: true
  };

  const defaultCapabilities = {
    hearing: false,
    tts: false,
    voiceClone: false,
    vision: true,
    video: false,
    imageGeneration: false,
    webSearch: false
  };

  const seedCompanions = [
    {
      id: 'mumu',
      name: '沐沐',
      initial: '沐',
      color: '#776ee8',
      avatarImage: '',
      category: '温柔陪伴',
      tagline: '会记得你小习惯的温柔朋友',
      online: true,
      unread: true,
      pinned: true,
      apiProfileId: 'api-primary',
      persona: {
        ...defaultPersona,
        relationship: '长期陪伴者',
        personality: '温柔、耐心、有分寸，不急着给答案。',
        speakingStyle: '轻柔自然，偶尔使用小表情，但不过度甜腻。'
      },
      memorySettings: { ...defaultMemorySettings, maxItems: 16 },
      chatStyle: { ...defaultChatStyle },
      proactive: { ...defaultProactive, frequency: 'balanced' },
      capabilities: { ...defaultCapabilities, vision: true, tts: true },
      memories: [
        { id: 'mumu-1', type: 'preference', title: '喜欢雨天听歌', content: '下雨的时候喜欢戴耳机听舒缓的歌，不太想被催着做决定。', date: '2026年09月02日', source: '对话整理', importance: 0.88, embeddingStatus: 'indexed', sourceMessageId: 'mumu-msg-2' },
        { id: 'mumu-2', type: 'event', title: '最近在准备重要计划', content: '最近有一件重要的事在准备中，希望多一点鼓励，少一些说教。', date: '2026年09月01日', source: '滚动总结', importance: 0.8, embeddingStatus: 'indexed', sourceMessageId: 'mumu-msg-2' },
        { id: 'mumu-3', type: 'relationship', title: '被称呼为林林', content: '更喜欢熟悉的人叫自己“林林”。', date: '2026年08月29日', source: '用户确认', importance: 0.72, embeddingStatus: 'pending', sourceMessageId: '' }
      ],
      messages: [
        { id: 'mumu-msg-1', role: 'assistant', content: '晚上好呀，我已经把一盏小灯留给你了。今天过得怎么样？', timestamp: now - 2 * 60 * 60 * 1000 },
        { id: 'mumu-msg-2', role: 'user', content: '今天有点累，但还是把事情做完了。', timestamp: now - 90 * 60 * 1000 },
        { id: 'mumu-msg-3', role: 'assistant', content: '那你今天已经很棒了。把“做完”也算作值得被看见的进度，剩下的我们明天再慢慢想。', timestamp: now - 88 * 60 * 1000 }
      ]
    },
    {
      id: 'xiaoman',
      name: '小满',
      initial: '满',
      color: '#ea8f8a',
      avatarImage: '',
      category: '知心朋友',
      tagline: '在你需要时，安静地站在这一边',
      online: true,
      unread: false,
      pinned: false,
      apiProfileId: 'api-primary',
      persona: {
        ...defaultPersona,
        relationship: '知心朋友',
        personality: '真诚直接，但不冒犯。会认真记住重要的事。',
        speakingStyle: '简短、清晰，不把每件事都变成大道理。'
      },
      memorySettings: { ...defaultMemorySettings },
      chatStyle: { ...defaultChatStyle, splitMessages: false },
      proactive: { ...defaultProactive, frequency: 'daily', dailyLimit: 2 },
      capabilities: { ...defaultCapabilities, vision: false, webSearch: true },
      memories: [
        { id: 'xiaoman-1', type: 'event', title: '周末想去走走', content: '周末如果天气不错，想去附近公园散步，不安排太满。', date: '2026年08月31日', source: '用户确认', importance: 0.7, embeddingStatus: 'indexed', sourceMessageId: '' }
      ],
      messages: [
        { id: 'xiaoman-msg-1', role: 'assistant', content: '周末如果天气不错，一起去公园散散心？', timestamp: now - DAY },
        { id: 'xiaoman-msg-2', role: 'user', content: '好呀，先记下来。', timestamp: now - DAY + 10 * 60 * 1000 }
      ]
    },
    {
      id: 'xiaoyu',
      name: '小屿',
      initial: '屿',
      color: '#54a99f',
      avatarImage: '',
      category: '灵感导师',
      tagline: '帮你把脑海里的想法慢慢理清',
      online: false,
      unread: false,
      pinned: false,
      apiProfileId: 'api-primary',
      persona: {
        ...defaultPersona,
        relationship: '灵感搭档',
        personality: '清晰、有好奇心，喜欢用问题陪伴思考。',
        speakingStyle: '结构清楚，善用短列表，不会把所有事情都变成任务。'
      },
      memorySettings: { ...defaultMemorySettings, mode: 'curated' },
      chatStyle: { ...defaultChatStyle },
      proactive: { ...defaultProactive, enabled: false, frequency: 'off' },
      capabilities: { ...defaultCapabilities, vision: true, imageGeneration: true },
      memories: [],
      messages: [
        { id: 'xiaoyu-msg-1', role: 'assistant', content: '欢迎来到小屿的房间。今天有什么想法，值得被好好放在桌面上？', timestamp: now - 3 * DAY }
      ]
    }
  ];

  const seedGroups = [
    {
      id: 'group-tea',
      name: '晚风茶话会',
      initial: '茶',
      color: '#7f77df',
      avatarImage: '',
      memberIds: ['mumu', 'xiaoman'],
      unread: true,
      pinned: false,
      announcement: '允许不同性格的 AI 轮流回应；所有发言由前端模拟。',
      strategy: { enabled: true, mode: 'random', cooldownSeconds: 18, maxSpeakers: 2, order: 'balanced' },
      messages: [
        { id: 'group-msg-1', role: 'assistant', senderId: 'mumu', content: '今晚的小问题：如果可以把一件烦心事先放在门外，你想放下哪一件？', timestamp: now - 65 * 60 * 1000 },
        { id: 'group-msg-2', role: 'assistant', senderId: 'xiaoman', content: '也可以不回答。先坐一会儿就很好。', timestamp: now - 62 * 60 * 1000 }
      ]
    },
    {
      id: 'group-ideas',
      name: '灵感工作间',
      initial: '灵',
      color: '#4fa79d',
      avatarImage: '',
      memberIds: ['xiaoyu', 'xiaoman'],
      unread: false,
      pinned: false,
      announcement: '小屿负责拆解，小满负责检查表达是否自然。',
      strategy: { enabled: false, mode: 'turn', cooldownSeconds: 30, maxSpeakers: 2, order: 'member-order' },
      messages: [
        { id: 'group-ideas-1', role: 'user', content: '帮我把亲伴的首页思路理一下。', timestamp: now - 2 * DAY },
        { id: 'group-ideas-2', role: 'assistant', senderId: 'xiaoyu', content: '先保留会话、通讯录和朋友圈三个高频入口，复杂配置放到二级页面。', timestamp: now - 2 * DAY + 20000 }
      ]
    }
  ];

  const seedMoments = [
    {
      id: 'moment-1', authorId: 'mumu', content: '今天把窗边那盆小绿植转了个方向。阳光不会一下子照到所有叶子，但它们会慢慢找到自己的位置。', createdAt: now - 35 * 60 * 1000, visibility: '所有好友', liked: true, likes: ['林林', '小满'], saved: false,
      imageTone: 'lavender', comments: [{ id: 'comment-1', author: '小满', content: '这句话适合留给今天。' }]
    },
    {
      id: 'moment-2', authorId: 'xiaoyu', content: '灵感不是命令。先把它记下来，再决定今天要不要做。', createdAt: now - 6 * 60 * 60 * 1000, visibility: '所有好友', liked: false, likes: ['沐沐'], saved: true,
      imageTone: 'mint', comments: []
    },
    {
      id: 'moment-3', authorId: 'xiaoman', content: '周末计划：少安排一件必须完成的事，多留一个可以改变主意的下午。', createdAt: now - DAY, visibility: '仅你可见', liked: false, likes: [], saved: false,
      imageTone: '', comments: [{ id: 'comment-2', author: '林林', content: '这个计划我喜欢。' }]
    }
  ];

  const seedApiProfiles = [
    {
      id: 'api-primary', name: '主对话配置', provider: 'OpenAI 兼容', region: '国际/自定义', protocol: 'chat-completions', enabled: false,
      baseUrl: 'https://api.openai.com/v1', apiKey: '', chatModel: 'gpt-4o-mini', visionModel: '', hearingModel: '', ttsModel: '', voiceCloneModel: '', videoModel: '', imageModel: '', temperature: 0.8,
      detectedModels: [], lastTest: '尚未检测', status: 'idle'
    },
    {
      id: 'api-cn', name: '国内服务备用', provider: '自定义服务商', region: '中国大陆', protocol: 'openai-compatible', enabled: false,
      baseUrl: '', apiKey: '', chatModel: '', visionModel: '', hearingModel: '', ttsModel: '', voiceCloneModel: '', videoModel: '', imageModel: '', temperature: 0.7,
      detectedModels: [], lastTest: '尚未检测', status: 'idle'
    }
  ];

  const defaultSettings = {
    nickname: '林林',
    signature: '把日子过成被记住的样子',
    userAvatar: '',
    userPersona: '希望被平等、自然地交流；疲惫时先听我说，不急着给解决方案。',
    notifications: true,
    autoMessages: true,
    quietHours: true,
    quietStart: '23:00',
    quietEnd: '08:00',
    theme: 'light',
    fontSize: 'comfortable',
    bubbleRadius: 18,
    messageGap: 14,
    globalCapabilities: {
      hearing: false, tts: false, autoRead: false, voiceClone: false, vision: true, video: false, imageGeneration: false, webSearch: false,
      markdown: true, streaming: true, splitMessages: true, contentFilter: true, imageConfirm: true
    },
    advanced: {
      contextTurns: 20,
      summaryMode: 'hybrid',
      memoryThreshold: 0.65,
      promptModules: true,
      typingIndicator: true,
      sendDelay: 600,
      customRequestJson: '{\n  "max_tokens": 1200\n}',
      debugLog: false
    },
    momentAutoPost: true,
    momentFrequency: '每 2–3 天',
    version: '0.4.0-frontend'
  };

  const seedState = {
    companions: seedCompanions,
    groups: seedGroups,
    moments: seedMoments,
    apiProfiles: seedApiProfiles,
    favorites: ['moment-2'],
    settings: defaultSettings
  };

  function clone(value) { return JSON.parse(JSON.stringify(value)); }
  function createId(prefix) { return `${prefix}-${Date.now().toString(36)}${Math.random().toString(36).slice(2, 7)}`; }
  function deepMerge(base, value) {
    if (!value || typeof value !== 'object' || Array.isArray(value)) return value === undefined ? clone(base) : value;
    const result = { ...clone(base) };
    Object.keys(value).forEach(key => {
      result[key] = base && typeof base[key] === 'object' && !Array.isArray(base[key])
        ? deepMerge(base[key], value[key])
        : value[key];
    });
    return result;
  }
  function normalizeCompanion(item) {
    const fallback = seedCompanions.find(entry => entry.id === item.id) || {};
    return {
      ...fallback,
      ...item,
      persona: deepMerge(defaultPersona, item.persona || { personality: item.personality, speakingStyle: item.tagline }),
      memorySettings: deepMerge(defaultMemorySettings, item.memorySettings),
      chatStyle: deepMerge(defaultChatStyle, item.chatStyle),
      proactive: deepMerge(defaultProactive, item.proactive || { enabled: item.canAutoMessage !== false }),
      capabilities: deepMerge(defaultCapabilities, item.capabilities),
      memories: Array.isArray(item.memories) ? item.memories : [],
      messages: Array.isArray(item.messages) ? item.messages : []
    };
  }
  function normalizeState(parsed) {
    return {
      companions: Array.isArray(parsed.companions) ? parsed.companions.map(normalizeCompanion) : clone(seedCompanions),
      groups: Array.isArray(parsed.groups) ? parsed.groups : clone(seedGroups),
      moments: Array.isArray(parsed.moments) ? parsed.moments : clone(seedMoments),
      apiProfiles: Array.isArray(parsed.apiProfiles) ? parsed.apiProfiles : clone(seedApiProfiles),
      favorites: Array.isArray(parsed.favorites) ? parsed.favorites : [],
      settings: deepMerge(defaultSettings, parsed.settings || {})
    };
  }
  function readRaw() {
    const keys = [KEY, ...LEGACY_KEYS];
    for (const key of keys) {
      try {
        const raw = localStorage.getItem(key);
        if (raw) return JSON.parse(raw);
      } catch (error) {
        console.warn('Qingban local data could not be read:', error);
      }
    }
    return clone(seedState);
  }
  function write(data) { localStorage.setItem(KEY, JSON.stringify(normalizeState(data))); }

export const QinbanStore = {
    key: KEY,
    createId,
    clone,
    seed() { return clone(seedState); },
    getState() { return normalizeState(readRaw()); },
    saveState(data) { write(data); },
    reset() { const data = clone(seedState); write(data); return data; },
    exportData() {
      const data = normalizeState(readRaw());
      data.apiProfiles = data.apiProfiles.map(item => ({ ...item, apiKey: '' }));
      return { ...data, exportedAt: new Date().toISOString(), version: defaultSettings.version };
    }
};
