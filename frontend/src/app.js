import { nextTick } from 'vue'
import { QinbanStore } from './store.js'

const ICON_PATHS = {
  message: '<path d="M5 5.5h14A2.5 2.5 0 0 1 21.5 8v7a2.5 2.5 0 0 1-2.5 2.5H11L5 21v-3.5A2.5 2.5 0 0 1 2.5 15V8A2.5 2.5 0 0 1 5 5.5Z"/><path d="M8 10h8M8 13h5"/>',
  users: '<path d="M16 19v-1.5a3.5 3.5 0 0 0-3.5-3.5h-5A3.5 3.5 0 0 0 4 17.5V19"/><circle cx="10" cy="7.5" r="3.5"/><path d="M16 7.5a3 3 0 0 1 0 5.8M18.5 14.5a3.5 3.5 0 0 1 2.5 3.3V19"/>',
  sparkle: '<path d="m12 3 1.25 4.3L17.5 9l-4.25 1.7L12 15l-1.25-4.3L6.5 9l4.25-1.7z"/><path d="m19 14 .6 2.1L21.5 17l-1.9.9L19 20l-.6-2.1-1.9-.9 1.9-.9zM5 15l.45 1.55L7 17l-1.55.45L5 19l-.45-1.55L3 17l1.55-.45z"/>',
  brain: '<path d="M9.5 4.5A3.5 3.5 0 0 0 6 8v.4a3.4 3.4 0 0 0-1.5 6.1A3.5 3.5 0 0 0 8 18h1.5M14.5 4.5A3.5 3.5 0 0 1 18 8v.4a3.4 3.4 0 0 1 1.5 6.1A3.5 3.5 0 0 1 16 18h-1.5M9.5 4.5v13M14.5 4.5v13M9.5 9h5M9.5 13h5"/>',
  settings: '<path d="M12 3.5v2M12 18.5v2M4.5 12h-2M21.5 12h-2M6 6 4.6 4.6M19.4 19.4 18 18M18 6l1.4-1.4M4.6 19.4 6 18"/><circle cx="12" cy="12" r="4.5"/>',
  moments: '<circle cx="12" cy="12" r="8.5"/><circle cx="12" cy="12" r="2.2" fill="currentColor" stroke="none"/>',
  home: '<path d="m3.5 11 8.5-7 8.5 7M5.5 10v10h13V10M9.5 20v-6h5v6"/>',
  palette: '<path d="M12 3.5a8.5 8.5 0 0 0 0 17h1.4a1.8 1.8 0 0 0 .8-3.4 1.8 1.8 0 0 1 .8-3.4H17a3.5 3.5 0 0 0 3.5-3.5A6.7 6.7 0 0 0 12 3.5Z"/><circle cx="7.5" cy="10" r=".8"/><circle cx="14" cy="6.5" r=".8"/>',
  search: '<circle cx="10.8" cy="10.8" r="6.3"/><path d="m16 16 4.2 4.2"/>', plus: '<path d="M12 5v14M5 12h14"/>',
  edit: '<path d="m4 16.7-.7 3.3 3.3-.7L18.4 7.5a2.2 2.2 0 0 0-3.1-3.1zM14.2 5.8l3.1 3.1"/>', close: '<path d="m6 6 12 12M18 6 6 18"/>', check: '<path d="m5 12 4.2 4.2L19 6.5"/>',
  'arrow-left': '<path d="M19 12H5M11 18l-6-6 6-6"/>', 'arrow-right': '<path d="M5 12h14M13 6l6 6-6 6"/>', send: '<path d="m21 3-7.2 18-3.8-7L3 10.2zM10 14 21 3"/>',
  heart: '<path d="M20.8 8.9c0 5.2-8.8 10.2-8.8 10.2S3.2 14.1 3.2 8.9A4.5 4.5 0 0 1 12 6.7a4.5 4.5 0 0 1 8.8 2.2Z"/>',
  clock: '<circle cx="12" cy="12" r="8.5"/><path d="M12 7v5l3.5 2"/>', trash: '<path d="M4.5 7h15M9 7V4.5h6V7M7 7l.8 12.5h8.4L17 7M10 10.5v6M14 10.5v6"/>',
  bell: '<path d="M18 9.5a6 6 0 0 0-12 0c0 7-2.5 7-2.5 8.5h17C20.5 16.5 18 16.5 18 9.5ZM10 21h4"/>',
  database: '<ellipse cx="12" cy="5.5" rx="7.5" ry="3"/><path d="M4.5 5.5v12c0 1.7 3.4 3 7.5 3s7.5-1.3 7.5-3v-12M4.5 11.5c0 1.7 3.4 3 7.5 3s7.5-1.3 7.5-3"/>',
  vector: '<circle cx="5" cy="6" r="2"/><circle cx="19" cy="7" r="2"/><circle cx="8" cy="18" r="2"/><circle cx="17" cy="17" r="2"/><path d="m7 7 10 0M6 8l2 8M18 9l-1 6M10 18h5M7 7l8 9"/>',
  download: '<path d="M12 4v10M8 10l4 4 4-4M5 19.5h14"/>', upload: '<path d="M12 20V10M8 14l4-4 4 4M5 4.5h14"/>', refresh: '<path d="M19 8.5A7.5 7.5 0 1 0 20 14M19 4v4.5h-4.5"/>',
  shield: '<path d="M12 3.5 19 6v5.2c0 4.2-2.8 7.4-7 9.3-4.2-1.9-7-5.1-7-9.3V6zM8.5 12l2.2 2.2 4.8-4.8"/>', info: '<circle cx="12" cy="12" r="9"/><path d="M12 10v6M12 7.2v.1"/>',
  paperclip: '<path d="m20 11.5-8.2 8.2a5 5 0 0 1-7.1-7.1l8.6-8.6a3.5 3.5 0 0 1 5 5L9.7 17.6a2 2 0 0 1-2.8-2.8l8-8"/>',
  voice: '<path d="M4 10v4h3l4 4V6l-4 4H4Z"/><path d="M15 9.5a4 4 0 0 1 0 5M17.5 7a7 7 0 0 1 0 10"/>',
  keyboard: '<rect x="3" y="6.5" width="18" height="11" rx="2"/><path d="M6.5 10h.1M9.5 10h.1M12.5 10h.1M15.5 10h.1M18 10h.1M6.5 13.5h11"/>',
  smile: '<circle cx="12" cy="12" r="9"/><path d="M8.5 14.5a4.5 4.5 0 0 0 7 0M8.5 9.5h.1M15.5 9.5h.1"/>',
  image: '<rect x="3.5" y="4.5" width="17" height="15" rx="2"/><circle cx="9" cy="10" r="2"/><path d="m5.5 17 4.2-4.2 3.2 3.2 2.1-2.1 3.5 3.1"/>', mic: '<rect x="9" y="3" width="6" height="11" rx="3"/><path d="M6 11a6 6 0 0 0 12 0M12 17v4M9 21h6"/>',
  eye: '<path d="M3 12s3.5-5.5 9-5.5S21 12 21 12s-3.5 5.5-9 5.5S3 12 3 12Z"/><circle cx="12" cy="12" r="2.5"/>', video: '<rect x="3" y="6" width="13" height="12" rx="2"/><path d="m16 10 5-3v10l-5-3z"/>',
  globe: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/>', code: '<path d="m8.5 8-4 4 4 4M15.5 8l4 4-4 4M14 5l-4 14"/>',
  chart: '<path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>', bookmark: '<path d="M6 4.5h12v16L12 17l-6 3.5z"/>', comment: '<path d="M4 5.5h16v11H9l-5 4z"/>',
  group: '<circle cx="9" cy="8" r="3"/><circle cx="17" cy="9" r="2.5"/><path d="M3.5 20c.4-4 2.4-6 5.5-6s5.1 2 5.5 6M14 14.5c3.5-.7 6 1.2 6.5 4.5"/>', user: '<circle cx="12" cy="8" r="4"/><path d="M4.5 21c.5-5 3-7 7.5-7s7 2 7.5 7"/>',
  lock: '<rect x="5" y="10" width="14" height="10" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/>', sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2"/>', moon: '<path d="M20 15.2A8.5 8.5 0 0 1 8.8 4 8.5 8.5 0 1 0 20 15.2Z"/>', more: '<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>'
};
function icon(name) { return `<svg viewBox="0 0 24 24" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${ICON_PATHS[name] || ICON_PATHS.sparkle}</g></svg>`; }
function todayLabel() { return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date()); }

const ChatView = {
  props: { thread: Object, companions: Array, messages: Array, isReplying: Boolean, inputMessage: String, apiActive: Boolean },
  emits: ['update-input', 'send', 'quick', 'proactive', 'back', 'open-settings', 'notice'],
  data() { return { voiceMode: false, showEmojiPanel: false, showMorePanel: false }; },
  methods: {
    icon,
    avatarStyle(item) { return { backgroundColor: item?.color || '#776ee8', backgroundImage: item?.avatarImage ? `url("${item.avatarImage}")` : 'none' }; },
    sender(message) { if (message.role === 'user') return { name: '我', initial: '我', color: '#776ee8' }; return this.thread.type === 'group' && message.senderId ? (this.companions.find(item => item.id === message.senderId) || this.thread) : this.thread; },
    formatMessageTime(timestamp) { return new Date(timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }); },
    toggleVoiceMode() { this.voiceMode = !this.voiceMode; this.showEmojiPanel = false; this.showMorePanel = false; },
    togglePanel(panel) { this.showEmojiPanel = panel === 'emoji' ? !this.showEmojiPanel : false; this.showMorePanel = panel === 'more' ? !this.showMorePanel : false; },
    insertEmoji(emoji) { this.$emit('update-input', `${this.inputMessage || ''}${emoji}`); this.showEmojiPanel = false; },
    notify(message) { this.$emit('notice', message); }
  },
  template: `<div class="chat-view">
    <div class="chat-contact-bar"><div class="chat-contact-info"><button class="chat-back-mobile" @click="$emit('back')"><span v-html="icon('arrow-left')"></span></button><div class="avatar" :style="avatarStyle(thread)"><span v-if="!thread.avatarImage">{{ thread.initial || thread.name.slice(0,1) }}</span></div><div><strong>{{ thread.name }}</strong><span :class="{'is-typing':isReplying}" aria-live="polite"><i class="status-dot"></i>{{ isReplying ? '正在输入…' : (thread.type === 'group' ? thread.memberIds.length + ' 位 AI 成员' : (thread.online ? '正在亲伴里' : '暂时安静')) }}</span></div></div><div class="chat-contact-actions"><button class="ghost-action" @click="$emit('proactive')"><span v-html="icon('send')"></span><span class="desktop-only">{{ thread.type === 'group' ? '触发一轮群聊' : '让 TA 主动说句话' }}</span><span class="mobile-only">主动消息</span></button><button class="round-action" @click="$emit('open-settings')"><span v-html="icon('settings')"></span></button></div></div>
    <div class="chat-memory-hint"><span v-html="icon(thread.type === 'group' ? 'group' : 'brain')"></span><span>{{ thread.type === 'group' ? '多 AI 发言调度为前端模拟，等待后端实现' : '长期记忆会按角色设置注入上下文' }}</span><button @click="$emit('open-settings')">查看配置 <span v-html="icon('arrow-right')"></span></button></div>
     <div class="chat-messages"><div v-if="!messages.length" class="chat-empty"><div class="empty-orb"><span v-html="icon('message')"></span></div><strong>从一句自然的话开始</strong><p>这是可交互前端原型。真实模型、记忆检索和主动消息需要后端接入。</p><button class="secondary-action" @click="$emit('quick','今天过得有点复杂，想和你聊聊。')">发个招呼</button></div><div v-for="(message,index) in messages" :key="message.id" class="message-block" :class="{mine:message.role==='user'}"><div v-if="index===0" class="date-divider">今天</div><div class="message-line"><div v-if="message.role!=='user'" class="message-avatar avatar" :style="avatarStyle(sender(message))"><span v-if="!sender(message).avatarImage">{{ sender(message).initial || sender(message).name.slice(0,1) }}</span></div><div class="message-body"><small v-if="thread.type==='group'&&message.role!=='user'" class="message-sender">{{ sender(message).name }}</small><div class="message-bubble">{{ message.content }}</div><time>{{ formatMessageTime(message.timestamp) }}<span v-if="message.proactive" class="proactive-tag">主动消息</span></time></div></div></div></div>
     <div class="chat-quick-replies"><button class="quick-chip" @click="$emit('quick','今天想轻松聊一会儿')">轻松聊聊</button><button class="quick-chip" @click="$emit('quick','帮我整理一下现在的心情')">整理心情</button><button class="quick-chip" @click="$emit('quick','你还记得我最近在忙什么吗？')">问问记忆</button></div>
     <div class="chat-composer wechat-composer" :class="{'voice-mode':voiceMode}">
       <button class="composer-icon composer-voice" :class="{active:voiceMode}" :aria-label="voiceMode?'切换文字输入':'切换语音输入'" @click="toggleVoiceMode"><span v-html="icon(voiceMode?'keyboard':'voice')"></span></button>
       <button v-if="voiceMode" class="hold-to-talk" @click="notify('语音输入仅为界面演示，暂未接入麦克风')">按住说话</button>
       <div v-else class="composer-input"><textarea rows="1" :value="inputMessage" @input="$emit('update-input',$event.target.value)" @keydown.enter.exact.prevent="$emit('send')" placeholder="输入消息…"></textarea><span>{{inputMessage.length}}/500</span><button v-if="inputMessage.trim()" class="inline-send" aria-label="发送消息" @click="$emit('send')"><span v-html="icon('send')"></span></button><button v-else class="inline-mic" aria-label="语音输入" @click="notify('语音输入仅为界面演示，暂未接入麦克风')"><span v-html="icon('mic')"></span></button></div>
       <div v-if="showEmojiPanel" class="composer-popover emoji-popover"><button v-for="emoji in ['😊','🙂','🤍','🌙','👍','✨','😂','🥺']" :key="emoji" @click="insertEmoji(emoji)">{{emoji}}</button></div>
       <div v-if="showMorePanel" class="composer-popover more-popover"><button @click="notify('图片功能等待后端接入')"><span v-html="icon('image')"></span>图片</button><button @click="notify('文件功能等待后端接入')"><span v-html="icon('paperclip')"></span>文件</button><button @click="notify('更多能力等待后端接入')"><span v-html="icon('sparkle')"></span>更多</button></div>
       <button class="composer-icon" aria-label="表情" :class="{active:showEmojiPanel}" @click="togglePanel('emoji')"><span v-html="icon('smile')"></span></button><button class="composer-icon" aria-label="更多" :class="{active:showMorePanel}" @click="togglePanel('more')"><span v-html="icon('plus')"></span></button>
     </div>
     <div class="chat-disclaimer"><span v-html="icon('shield')"></span>{{apiActive?'已选择 API 配置；浏览器直连仅用于联调':'当前使用本地模拟回复，不会产生 API 消耗'}}</div>
  </div>`
};

const initialState = QinbanStore.getState();
export const appOptions = {
  components: { ChatView },
  data() { return {
    ...initialState, isBooting: true, activeView: 'inbox', previousView: 'inbox', activeChatRef: null, addMenuOpen: false, selectedCompanionId: initialState.companions[0]?.id || '', selectedApiId: initialState.apiProfiles[0]?.id || '', aiDetailTab: 'profile', settingsTab: 'api', memoryLabTab: 'records', inboxFilter: 'all', contactFilter: 'all', searchQuery: '', memorySearchQuery: '最近的重要计划', memorySearchResults: [], inputMessage: '', isReplying: false, modelDetecting: false,
    showGroupDialog: false, showProfileDialog: false, showMemoryDialog: false, showMomentDialog: false, showAbout: false, utilityPanel: '', editingMemory: null, companionForm: {}, groupForm: {}, profileForm: {}, memoryForm: {}, momentForm: {content:'',visibility:'所有好友',imageTone:''}, commentDrafts: {}, toastMessage: '', toastTimer: null, persistTimer: null,
    providerOptions: [
      {name:'OpenAI 兼容',region:'国际/自定义',baseUrl:'https://api.openai.com/v1'}, {name:'DeepSeek',region:'中国大陆',baseUrl:'https://api.deepseek.com'}, {name:'阿里云百炼',region:'中国大陆',baseUrl:''}, {name:'火山方舟',region:'中国大陆',baseUrl:''}, {name:'腾讯混元',region:'中国大陆',baseUrl:''}, {name:'智谱 AI',region:'中国大陆',baseUrl:''}, {name:'Anthropic',region:'国际',baseUrl:''}, {name:'Google AI',region:'国际',baseUrl:''}, {name:'OpenRouter',region:'国际聚合',baseUrl:''}, {name:'Ollama / 本地',region:'本机',baseUrl:'http://localhost:11434/v1'}, {name:'自定义服务商',region:'自定义',baseUrl:''}
    ],
    capabilityItems: [
      {key:'hearing',icon:'mic',title:'听觉 / 语音识别',description:'语音转文字、语音消息理解'}, {key:'tts',icon:'bell',title:'语音合成',description:'生成语音条或自动朗读回复'}, {key:'autoRead',icon:'send',title:'自动朗读',description:'回复完成后自动播放语音'}, {key:'voiceClone',icon:'user',title:'声音复刻',description:'绑定后端声音授权与复刻任务'}, {key:'vision',icon:'eye',title:'视觉识图',description:'理解用户发送的图片内容'}, {key:'video',icon:'video',title:'视频理解',description:'上传视频后的理解与问答'}, {key:'imageGeneration',icon:'image',title:'自动生图 / 文生图',description:'按对话或明确指令生成图片'}, {key:'webSearch',icon:'globe',title:'联网搜索',description:'由后端搜索工具获取最新信息'}, {key:'markdown',icon:'code',title:'Markdown',description:'显示列表、代码与强调格式'}, {key:'streaming',icon:'message',title:'流式输出',description:'逐步显示模型返回内容'}, {key:'splitMessages',icon:'more',title:'多条消息拆分',description:'把长回复拆成自然的多条消息'}, {key:'contentFilter',icon:'shield',title:'内容安全过滤',description:'后端审核、危机提示与边界控制'}, {key:'imageConfirm',icon:'check',title:'生图二次确认',description:'执行图片生成前由用户确认'}
    ],
    navItems: [{id:'inbox',label:'会话',icon:'message'},{id:'contacts',label:'通讯录',icon:'users'},{id:'moments',label:'朋友圈',icon:'moments'},{id:'mine',label:'我的',icon:'user'}]
  }; },
  computed: {
    mobileNavItems() { return this.navItems; },
    userInitial() { return (this.settings.nickname || '你').slice(0,1); },
    activeCompanion() { return this.companions.find(item => item.id === this.selectedCompanionId) || this.companions[0] || null; },
    activeApiProfile() { return this.apiProfiles.find(item => item.id === this.selectedApiId) || this.apiProfiles[0] || null; },
    allThreads() {
      const a = this.companions.map(item => ({...item,type:'companion',threadId:item.id,lastMessage:this.lastMessage(item.messages),lastTimestamp:this.lastTimestamp(item.messages)}));
      const b = this.groups.map(item => ({...item,type:'group',threadId:item.id,lastMessage:this.lastMessage(item.messages),lastTimestamp:this.lastTimestamp(item.messages)}));
      return [...a,...b].sort((x,y) => Number(y.pinned)-Number(x.pinned) || y.lastTimestamp-x.lastTimestamp);
    },
    filteredThreads() {
      const q = this.searchQuery.trim().toLowerCase();
      return this.allThreads.filter(item => (this.inboxFilter==='all'||(this.inboxFilter==='unread'&&item.unread)||this.inboxFilter===item.type) && (!q||item.name.toLowerCase().includes(q)||item.lastMessage.toLowerCase().includes(q)));
    },
    activeThread() {
      if (!this.activeChatRef) return null;
      const item = (this.activeChatRef.type==='group'?this.groups:this.companions).find(entry => entry.id===this.activeChatRef.id);
      return item ? {...item,type:this.activeChatRef.type} : null;
    },
    activeMessages() { return this.activeThread?.messages || []; },
    activeChatApiReady() {
      if (!this.activeThread || this.activeThread.type==='group') return false;
      const p = this.apiProfiles.find(item => item.id===this.activeThread.apiProfileId);
      return Boolean(p?.enabled && p.baseUrl?.trim() && p.chatModel?.trim());
    },
    unreadTotal() { return this.allThreads.reduce((n,item)=>n+(item.unread?1:0),0); },
    memoryCount() { return this.companions.reduce((n,item)=>n+(item.memories||[]).length,0); },
    allMemories() { return this.companions.flatMap(c=>(c.memories||[]).map(m=>({...m,companion:c}))).sort((a,b)=>this.memoryDateValue(b.date)-this.memoryDateValue(a.date)); },
    filteredContacts() {
      const q=this.searchQuery.trim().toLowerCase();
      return this.companions.filter(item => (this.contactFilter==='all'||item.category===this.contactFilter||(this.contactFilter==='online'&&item.online)) && (!q||`${item.name}${item.category}${item.tagline}`.toLowerCase().includes(q)));
    },
    contactCategories() { return ['all',...new Set(this.companions.map(item=>item.category))]; },
    proactiveCount() { return this.companions.filter(item=>item.proactive?.enabled).length; },
    enabledCapabilityCount() { return Object.values(this.settings.globalCapabilities||{}).filter(Boolean).length; },
    viewTitle() { return ({inbox:'会话',chat:this.activeThread?.name||'聊天',contacts:'通讯录',moments:'朋友圈',studio:'AI 与群聊','ai-detail':this.companionForm?.name?`${this.companionForm.name} · 详细设置`:'创建 AI 好友','memory-lab':'长期记忆与数据台',mine:'我的',settings:'设置中心',usage:'API 消耗诊断'})[this.activeView]||'亲伴'; },
    viewSubtitle() { return ({inbox:'和好友聊聊正在发生的事。',contacts:'找到你的 AI 好友，直接开始聊天。',moments:'看看好友的近况，也分享你的生活。',studio:'创建角色与 AI 群聊，查看后端待接入项。','memory-lab':'面向后端的数据结构、向量检索与上下文演示。',mine:'个人资料、收藏和更多设置。',settings:'API、模型能力、外观与高级参数统一配置。',usage:'前端估算，不代表真实供应商账单。'})[this.activeView]||''; },
    isSecondaryView() { return ['ai-detail','settings','usage'].includes(this.activeView); },
    totalMessages() { return this.companions.reduce((n,i)=>n+(i.messages||[]).length,0)+this.groups.reduce((n,i)=>n+(i.messages||[]).length,0); },
    personaChars() { return this.companions.reduce((n,i)=>n+JSON.stringify(i.persona||{}).length,0); },
    estimatedTokens() { return Math.round(2800+this.totalMessages*560+this.memoryCount*210+this.settings.advanced.contextTurns*45); },
    estimatedCost() { return (this.estimatedTokens/1000000*1.8).toFixed(4); },
    usageLevel() { const s=this.settings.advanced.contextTurns+this.memoryCount*2+(this.settings.advanced.summaryMode==='full'?24:0)+this.enabledCapabilityCount*2; return s>65?{label:'偏高',tone:'danger',score:78}:s>38?{label:'适中',tone:'warning',score:54}:{label:'较低',tone:'good',score:30}; },
    savedMoments() { return this.moments.filter(item=>item.saved||this.favorites.includes(item.id)); },
    myMoments() { return this.moments.filter(item=>item.authorId==='user'); },
    selectedGroupMembers() { return this.companions.filter(item=>this.groupForm.memberIds?.includes(item.id)); },
    backendContractPreview() { return JSON.stringify({companionId:this.activeCompanion?.id||'companion-id',query:this.memorySearchQuery||'用户当前消息',topK:8,threshold:Number(this.activeCompanion?.memorySettings?.searchThreshold||0.65),include:['episodic_memory','user_preference','relationship_event'],output:{memories:'MemoryHit[]',summary:'string',traceId:'string'}},null,2); }
  },
  watch: { settings:{deep:true,handler(){this.schedulePersist();}}, activeMessages(){this.$nextTick(this.scrollChatToBottom);}, activeView(){this.$nextTick(()=>{const area=document.querySelector('.content-area');if(area)area.scrollTop=0;this.scrollChatToBottom();});} },
  mounted() { window.setTimeout(()=>{this.isBooting=false;},1800); this.$nextTick(this.scrollChatToBottom); },
  methods: {
    icon,
    clone(v){return QinbanStore.clone(v);},
    stateSnapshot(){return{companions:this.companions,groups:this.groups,moments:this.moments,apiProfiles:this.apiProfiles,favorites:this.favorites,settings:this.settings};},
    persist(){QinbanStore.saveState(this.stateSnapshot());},
    schedulePersist(){if(this.persistTimer)clearTimeout(this.persistTimer);this.persistTimer=setTimeout(()=>this.persist(),180);},
    navigate(view){this.previousView=['chat','ai-detail','settings','usage'].includes(this.activeView)?this.previousView:this.activeView;this.activeView=view;this.searchQuery='';this.addMenuOpen=false;this.$nextTick(this.scrollChatToBottom);},
    toggleInboxAddMenu(){this.addMenuOpen=!this.addMenuOpen;},
    startFriendFromInbox(){this.addMenuOpen=false;this.openNewCompanion();},
    startGroupFromInbox(){this.addMenuOpen=false;this.openGroupDialog();},
    goBackSecondary(){this.activeView=this.previousView||'mine';},
    focusSearch(){this.$nextTick(()=>this.$refs.searchInput?.focus());},
    openThread(type,id){this.addMenuOpen=false;this.activeChatRef={type,id};const item=(type==='group'?this.groups:this.companions).find(x=>x.id===id);if(item)item.unread=false;this.persist();if(window.innerWidth<=760||this.activeView!=='inbox'){this.previousView=this.activeView;this.activeView='chat';}this.$nextTick(this.scrollChatToBottom);},
    closeChat(){this.activeView=this.previousView&&this.previousView!=='chat'?this.previousView:'inbox';this.inputMessage='';},
    lastTimestamp(messages){return messages?.length?Number(messages[messages.length-1].timestamp||0):0;},
    lastMessage(messages){return messages?.length?String(messages[messages.length-1].content||''):'还没有聊天记录，去打个招呼吧';},
    lastMessageTime(thread){const t=thread.lastTimestamp||this.lastTimestamp(thread.messages);if(!t)return'';const d=new Date(t),n=new Date();return d.toDateString()===n.toDateString()?d.toLocaleTimeString('zh-CN',{hour:'2-digit',minute:'2-digit'}):d.toLocaleDateString('zh-CN',{month:'numeric',day:'numeric'});},
    formatRelativeTime(t){const d=Date.now()-Number(t||0);if(d<3600000)return`${Math.max(1,Math.floor(d/60000))} 分钟前`;if(d<86400000)return`${Math.floor(d/3600000)} 小时前`;return`${Math.floor(d/86400000)} 天前`;},
    scrollChatToBottom(){document.querySelectorAll('.chat-messages').forEach(item=>{item.scrollTop=item.scrollHeight;});},
    avatarStyle(item){return{backgroundColor:item?.color||'#776ee8',backgroundImage:item?.avatarImage?`url("${item.avatarImage}")`:'none'};},
    userAvatarStyle(){return{backgroundImage:this.settings.userAvatar?`url("${this.settings.userAvatar}")`:'none'};},
    memberNames(group){return(group.memberIds||[]).map(id=>this.companions.find(x=>x.id===id)?.name).filter(Boolean).join('、')||'尚未添加成员';},
    capabilityNames(item){const map={hearing:'听觉',tts:'语音',voiceClone:'声音复刻',vision:'视觉',video:'视频',imageGeneration:'生图',webSearch:'联网'};return Object.entries(item.capabilities||{}).filter(([,v])=>v).map(([k])=>map[k]).filter(Boolean);},
    openNewCompanion(){this.previousView=this.activeView;this.selectedCompanionId='';this.companionForm=this.blankCompanion();this.aiDetailTab='profile';this.activeView='ai-detail';},
    openCompanionEditor(companion){this.previousView=this.activeView;this.selectedCompanionId=companion.id;this.companionForm=this.clone(companion);this.aiDetailTab='profile';this.activeView='ai-detail';},
    blankCompanion(){return{id:'',name:'',initial:'',color:'#776ee8',avatarImage:'',category:'温柔陪伴',tagline:'',online:true,unread:false,pinned:false,apiProfileId:this.apiProfiles[0]?.id||'',persona:{identity:'一位愿意长期陪伴、尊重边界的 AI 好友',relationship:'知心朋友',personality:'温柔、真诚、有耐心。',speakingStyle:'自然口语，先回应感受。',boundaries:'尊重用户选择，遇到高风险问题建议寻求现实帮助。',forbiddenTopics:'不诱导依赖，不冒充现实中的特定人物。'},memorySettings:{enabled:true,mode:'hybrid',summaryMode:'rolling',timeRangeDays:365,searchThreshold:.65,maxItems:12},chatStyle:{markdown:true,streaming:true,typing:true,splitMessages:true,replyDelay:650,bubbleStyle:'soft'},proactive:{enabled:true,start:'09:00',end:'22:30',frequency:'balanced',minMinutes:45,maxMinutes:240,dailyLimit:4,avoidBusyTime:true},capabilities:{hearing:false,tts:false,voiceClone:false,vision:false,video:false,imageGeneration:false,webSearch:false},memories:[],messages:[]};},
    saveCompanionDetails(){const f=this.companionForm;if(!f.name?.trim())return this.showToast('请填写 AI 好友名称');f.name=f.name.trim();f.initial=f.initial?.trim()||f.name.slice(0,1);if(!f.id){f.id=QinbanStore.createId('companion');f.messages=[{id:QinbanStore.createId('message'),role:'assistant',content:`你好，我是${f.name}。角色页面已经准备好，等后端接入后我们就能正式聊天。`,timestamp:Date.now()}];this.companions.push(this.clone(f));}else{const i=this.companions.findIndex(x=>x.id===f.id);if(i>=0)this.companions.splice(i,1,this.clone(f));}this.selectedCompanionId=f.id;this.persist();this.showToast('AI 好友配置已保存');},
    removeCompanion(companion){if(!companion||!confirm(`确定删除 ${companion.name} 吗？相关聊天和记忆也会从本机移除。`))return;this.companions=this.companions.filter(x=>x.id!==companion.id);this.groups.forEach(g=>g.memberIds=(g.memberIds||[]).filter(id=>id!==companion.id));this.activeChatRef=null;this.selectedCompanionId=this.companions[0]?.id||'';this.persist();this.addMenuOpen=false;this.activeView=this.previousView==='chat'?'inbox':(['inbox','contacts','studio'].includes(this.previousView)?this.previousView:'contacts');this.showToast('AI 好友已删除');},
    async handleImageUpload(event,target){const input=event.target,file=input.files?.[0];input.value='';if(!file)return;if(!file.type.startsWith('image/')||file.size>5*1024*1024)return this.showToast('请选择 5 MB 以内的图片');try{const image=await this.compressAvatar(file);if(target==='companion')this.companionForm.avatarImage=image;if(target==='group')this.groupForm.avatarImage=image;if(target==='user')this.profileForm.userAvatar=image;this.showToast('头像已更新，保存后生效');}catch(e){this.showToast('头像处理失败，请更换图片');}},
    compressAvatar(file){return new Promise((resolve,reject)=>{const r=new FileReader();r.onerror=reject;r.onload=()=>{const image=new Image();image.onerror=reject;image.onload=()=>{const size=320,c=document.createElement('canvas');c.width=size;c.height=size;const x=c.getContext('2d');if(!x)return reject();const scale=Math.max(size/image.width,size/image.height),w=image.width*scale,h=image.height*scale;x.imageSmoothingEnabled=true;x.imageSmoothingQuality='high';x.fillStyle='#eff1f8';x.fillRect(0,0,size,size);x.drawImage(image,(size-w)/2,(size-h)/2,w,h);resolve(c.toDataURL('image/jpeg',.9));};image.src=r.result;};r.readAsDataURL(file);});},
    openGroupDialog(group=null,preselectId=''){this.groupForm=group?this.clone(group):{id:'',name:'',initial:'',color:'#776ee8',avatarImage:'',memberIds:preselectId?[preselectId]:[],unread:false,pinned:false,announcement:'',strategy:{enabled:true,mode:'random',cooldownSeconds:18,maxSpeakers:2,order:'balanced'},messages:[]};this.showGroupDialog=true;},
    toggleGroupMember(id){const list=this.groupForm.memberIds||(this.groupForm.memberIds=[]),i=list.indexOf(id);i>=0?list.splice(i,1):list.push(id);},
    saveGroup(){if(!this.groupForm.name?.trim())return this.showToast('请填写群聊名称');if((this.groupForm.memberIds||[]).length<2)return this.showToast('请至少选择 2 位 AI 好友');const f=this.clone(this.groupForm);f.name=f.name.trim();f.initial=f.initial||f.name.slice(0,1);if(!f.id){f.id=QinbanStore.createId('group');f.messages=[{id:QinbanStore.createId('message'),role:'assistant',senderId:f.memberIds[0],content:'群聊框架已经创建。多 AI 主动发言与调度逻辑等待后端接入。',timestamp:Date.now()}];this.groups.push(f);}else{const i=this.groups.findIndex(x=>x.id===f.id);if(i>=0)this.groups.splice(i,1,f);}this.persist();this.showGroupDialog=false;this.showToast('AI 群聊已保存');},
    removeGroup(group){if(!group||!confirm(`确定删除群聊“${group.name}”吗？`))return;this.groups=this.groups.filter(x=>x.id!==group.id);if(this.activeChatRef?.type==='group'&&this.activeChatRef.id===group.id){this.activeChatRef=null;if(this.activeView==='chat')this.activeView='inbox';}this.persist();this.showGroupDialog=false;this.addMenuOpen=false;this.showToast('群聊已删除');},
    async sendMessage(){const text=this.inputMessage.trim();if(!text||!this.activeThread||this.isReplying)return;const target=(this.activeThread.type==='group'?this.groups:this.companions).find(x=>x.id===this.activeThread.id);if(!target)return;target.messages.push({id:QinbanStore.createId('message'),role:'user',content:text.slice(0,500),timestamp:Date.now()});this.inputMessage='';this.isReplying=true;this.persist();await nextTick();this.scrollChatToBottom();try{if(this.activeThread.type==='group')await this.generateGroupReplies(target,text);else{if(!this.activeChatApiReady){const delay=Math.min(Math.max(Number(target.chatStyle?.replyDelay||450),180),1500);await this.wait(delay);}const reply=this.activeChatApiReady?await this.requestApiReply(target,text):this.generateReply(text,target);target.messages.push({id:QinbanStore.createId('message'),role:'assistant',content:reply,timestamp:Date.now()});this.maybeCaptureMemory(text,target);}}catch(e){target.messages.push({id:QinbanStore.createId('message'),role:'assistant',content:this.generateReply(text,target),timestamp:Date.now(),fallback:true});this.showToast('API 联调未成功，已使用本地模拟回复');}finally{this.isReplying=false;this.persist();await nextTick();this.scrollChatToBottom();}},
    sendQuickMessage(text){this.inputMessage=text;this.sendMessage();},
    async generateGroupReplies(group,text){const ids=(group.memberIds||[]).slice(0,Math.max(1,Number(group.strategy?.maxSpeakers||2)));for(let i=0;i<ids.length;i++){await this.wait(i===0?650:420);const member=this.companions.find(x=>x.id===ids[i]);if(!member)continue;group.messages.push({id:QinbanStore.createId('message'),role:'assistant',senderId:member.id,content:i===0?this.generateReply(text,member):`${member.name}补充一句：我们也可以先把这件事拆小一点，不用一次处理完。`,timestamp:Date.now()});this.persist();await nextTick();this.scrollChatToBottom();}},
    generateReply(text,companion){const name=companion?.name||'我';if(/累|疲惫|困/.test(text))return'听起来你已经撑了很久。现在不必马上解决所有事，先允许自己停一小会儿。';if(/难过|不开心|委屈/.test(text))return'我在。你可以只讲发生了什么，也可以先不解释，我会陪你把这段情绪放稳一点。';if(/焦虑|担心|害怕/.test(text))return'我们先把“现在能做的一小步”和“暂时无法控制的部分”分开，好吗？';if(/记得|记忆/.test(text))return(companion.memories||[]).length?`我记得一些：${companion.memories.slice(0,2).map(x=>x.title).join('、')}。后端接入后会通过向量检索更准确地选取。`:'目前还没有确认过的长期记忆。你可以在记忆数据台添加或查看。';if(/你好|在吗|嗨/.test(text))return`我在，${this.settings.nickname||'你'}。今天想轻松聊聊，还是有件具体的事想说？`;return`${name}听见了。你更希望我先陪你把感受说清楚，还是一起想一个很小的下一步？`;},
    maybeCaptureMemory(text,companion){if(!companion.memorySettings?.enabled||!/喜欢|不喜欢|最近|以后|记住/.test(text)||text.length<8)return;companion.memories.unshift({id:QinbanStore.createId('memory'),type:'preference',title:'待确认的对话记忆',content:text.slice(0,120),date:todayLabel(),source:'前端规则模拟',importance:.58,embeddingStatus:'pending',sourceMessageId:companion.messages.at(-2)?.id||''});},
    wait(ms){return new Promise(resolve=>setTimeout(resolve,ms));},
    apiEndpoint(base,kind){base=String(base||'').trim().replace(/\/$/,'');if(!base)return'';if(kind==='models')return/\/models$/i.test(base)?base:`${base}/models`;return/\/chat\/completions$/i.test(base)?base:`${base}/chat/completions`;},
    async requestApiReply(companion){const p=this.apiProfiles.find(x=>x.id===companion.apiProfileId);if(!p)throw Error('No API profile');const memories=(companion.memories||[]).slice(0,companion.memorySettings?.maxItems||8).map(x=>`- ${x.title}: ${x.content}`).join('\n');const payload={model:p.chatModel,temperature:Number(p.temperature||.8),messages:[{role:'system',content:`你是亲伴中的 AI 好友“${companion.name}”。\n身份：${companion.persona.identity}\n关系：${companion.persona.relationship}\n性格：${companion.persona.personality}\n表达：${companion.persona.speakingStyle}\n边界：${companion.persona.boundaries}\n相关记忆：\n${memories||'暂无'}`},...(companion.messages||[]).slice(-Math.max(4,this.settings.advanced.contextTurns)).map(x=>({role:x.role,content:x.content}))]};Object.assign(payload,this.parseCustomJson());const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),30000);try{const r=await fetch(this.apiEndpoint(p.baseUrl,'chat'),{method:'POST',headers:{'Content-Type':'application/json',...(p.apiKey?{Authorization:`Bearer ${p.apiKey}`}:{})},body:JSON.stringify(payload),signal:controller.signal});if(!r.ok)throw Error(`HTTP ${r.status}`);const data=await r.json(),content=data.choices?.[0]?.message?.content||data.output_text||data.output?.text;if(!content)throw Error('Empty');return content;}finally{clearTimeout(timer);}},
    parseCustomJson(){try{return JSON.parse(this.settings.advanced.customRequestJson||'{}');}catch(e){return{};}},
    sendProactiveMessage(thread){if(!thread)return;if(thread.type==='group')return this.triggerGroupRound(thread);const target=this.companions.find(x=>x.id===thread.id);if(!target)return;target.messages.push({id:QinbanStore.createId('message'),role:'assistant',content:'刚刚想到你了。现在不用认真回答，只想问问：今天有没有一件小事值得被记住？',timestamp:Date.now(),proactive:true});this.persist();this.showToast('已生成一条前端模拟主动消息');this.$nextTick(this.scrollChatToBottom);},
    async triggerGroupRound(thread){const group=this.groups.find(x=>x.id===thread.id);if(!group)return;this.isReplying=true;await this.generateGroupReplies(group,'请大家主动聊一轮');this.isReplying=false;this.showToast('已模拟一轮多 AI 主动聊天');},
    openChatSettings(thread){thread.type==='group'?this.openGroupDialog(this.groups.find(x=>x.id===thread.id)):this.openCompanionEditor(this.companions.find(x=>x.id===thread.id));},
    openMemoryDialog(memory=null){this.editingMemory=memory;this.memoryForm=memory?{...this.clone(memory),companionId:memory.companion?.id||memory.companionId}:{id:'',companionId:this.activeCompanion?.id||this.companions[0]?.id||'',type:'preference',title:'',content:'',date:todayLabel(),source:'手动添加',importance:.7,embeddingStatus:'pending',sourceMessageId:''};this.showMemoryDialog=true;},
    saveMemory(){if(!this.memoryForm.companionId||!this.memoryForm.title?.trim()||!this.memoryForm.content?.trim())return this.showToast('请补充记忆标题和内容');const c=this.companions.find(x=>x.id===this.memoryForm.companionId);if(!c)return;const e={...this.clone(this.memoryForm),id:this.memoryForm.id||QinbanStore.createId('memory')},i=c.memories.findIndex(x=>x.id===e.id);i>=0?c.memories.splice(i,1,e):c.memories.unshift(e);this.persist();this.showMemoryDialog=false;this.showToast('记忆记录已保存');},
    removeMemory(memory){if(!confirm('确定删除这条长期记忆吗？'))return;const c=this.companions.find(x=>x.id===(memory.companion?.id||memory.companionId));if(!c)return;c.memories=c.memories.filter(x=>x.id!==memory.id);this.persist();this.showToast('记忆已删除');},
    memoryTypeLabel(type){return({preference:'用户偏好',event:'关系事件',relationship:'关系记忆',summary:'对话摘要'})[type]||'长期记忆';},
    memoryDateValue(date){return new Date(String(date||'').replace(/年|月/g,'-').replace(/日/g,'')).getTime()||0;},
    runMemorySearch(){const q=this.memorySearchQuery.trim();if(!q)return this.showToast('先输入一段检索内容');this.memorySearchResults=this.allMemories.slice(0,5).map((x,i)=>({...x,score:Math.max(.61,.91-i*.07-(q.includes(x.title.slice(0,2))?0:.03)).toFixed(2)}));this.showToast('已完成前端模拟检索');},
    openMomentDialog(){this.momentForm={content:'',visibility:'所有好友',imageTone:'lavender'};this.showMomentDialog=true;},
    publishMoment(){if(!this.momentForm.content.trim())return this.showToast('写点内容再发布吧');this.moments.unshift({id:QinbanStore.createId('moment'),authorId:'user',content:this.momentForm.content.trim(),createdAt:Date.now(),visibility:this.momentForm.visibility,liked:false,likes:[],saved:false,imageTone:this.momentForm.imageTone,comments:[]});this.persist();this.showMomentDialog=false;this.showToast('动态已发布到本地朋友圈');},
    simulateAiMoment(){const c=this.companions[Math.floor(Math.random()*Math.max(1,this.companions.length))];if(!c)return;const lines=['今天路过一小片很安静的光，想把它分享给你。','提醒自己：可以认真生活，也可以偶尔什么都不完成。','刚刚想起一段我们聊过的话。重要的不是记住全部，而是下次还能接上。'];this.moments.unshift({id:QinbanStore.createId('moment'),authorId:c.id,content:lines[Math.floor(Math.random()*lines.length)],createdAt:Date.now(),visibility:'所有好友',liked:false,likes:[],saved:false,imageTone:['lavender','mint','coral'][Math.floor(Math.random()*3)],comments:[]});this.persist();this.showToast(`${c.name} 发布了一条前端模拟动态`);},
    momentAuthor(moment){return moment.authorId==='user'?{name:this.settings.nickname,initial:this.userInitial,color:'#776ee8',avatarImage:this.settings.userAvatar}:this.companions.find(x=>x.id===moment.authorId)||{name:'亲伴好友',initial:'亲',color:'#776ee8'};},
    toggleMomentLike(moment){moment.liked=!moment.liked;const n=this.settings.nickname||'我';moment.likes=moment.likes||[];if(moment.liked&&!moment.likes.includes(n))moment.likes.push(n);if(!moment.liked)moment.likes=moment.likes.filter(x=>x!==n);this.persist();},
    toggleMomentSave(moment){moment.saved=!moment.saved;this.favorites=this.moments.filter(x=>x.saved).map(x=>x.id);this.persist();this.showToast(moment.saved?'已加入收藏':'已取消收藏');},
    addMomentComment(moment){const t=String(this.commentDrafts[moment.id]||'').trim();if(!t)return;moment.comments.push({id:QinbanStore.createId('comment'),author:this.settings.nickname||'我',content:t});this.commentDrafts[moment.id]='';this.persist();},
    editProfile(){this.profileForm={nickname:this.settings.nickname,signature:this.settings.signature,userAvatar:this.settings.userAvatar,userPersona:this.settings.userPersona};this.showProfileDialog=true;},
    saveProfile(){Object.assign(this.settings,this.profileForm);this.settings.nickname=this.settings.nickname.trim()||'你';this.persist();this.showProfileDialog=false;this.showToast('个人资料已保存');},
    openSettings(tab='api'){this.previousView=this.activeView;this.settingsTab=tab;this.activeView='settings';},
    openUsage(){this.previousView=this.activeView;this.activeView='usage';},
    openUtility(panel){this.utilityPanel=panel;},
    addApiProfile(){const p={id:QinbanStore.createId('api'),name:`配置 ${this.apiProfiles.length+1}`,provider:'自定义服务商',region:'自定义',protocol:'openai-compatible',enabled:false,baseUrl:'',apiKey:'',chatModel:'',visionModel:'',hearingModel:'',ttsModel:'',voiceCloneModel:'',videoModel:'',imageModel:'',temperature:.8,detectedModels:[],lastTest:'尚未检测',status:'idle'};this.apiProfiles.push(p);this.selectedApiId=p.id;this.persist();},
    selectProvider(profile,provider){profile.provider=provider.name;profile.region=provider.region;if(provider.baseUrl)profile.baseUrl=provider.baseUrl;profile.detectedModels=[];profile.status='idle';this.persist();},
    removeApiProfile(profile){if(this.apiProfiles.length<=1)return this.showToast('至少保留一套 API 配置');if(!confirm(`删除“${profile.name}”吗？`))return;this.apiProfiles=this.apiProfiles.filter(x=>x.id!==profile.id);this.selectedApiId=this.apiProfiles[0].id;this.companions.forEach(x=>{if(x.apiProfileId===profile.id)x.apiProfileId=this.selectedApiId;});this.persist();},
    async detectModels(profile){if(!profile)return;this.modelDetecting=true;profile.status='testing';profile.lastTest='正在识别…';const demo=[{id:'chat-model-demo',capabilities:['对话','流式']},{id:'vision-model-demo',capabilities:['视觉','图片理解']},{id:'speech-model-demo',capabilities:['听觉','语音']},{id:'image-model-demo',capabilities:['文生图']}];try{if(!profile.baseUrl?.trim()||(!profile.apiKey&&!/localhost|127\.0\.0\.1/.test(profile.baseUrl)))throw Error('demo');const r=await fetch(this.apiEndpoint(profile.baseUrl,'models'),{headers:profile.apiKey?{Authorization:`Bearer ${profile.apiKey}`}:{}});if(!r.ok)throw Error(`HTTP ${r.status}`);const data=await r.json(),models=Array.isArray(data.data)?data.data:Array.isArray(data.models)?data.models:[];profile.detectedModels=models.slice(0,30).map(x=>({id:x.id||x.name,capabilities:this.inferModelCapabilities(x.id||x.name)}));profile.status='success';profile.lastTest=`已识别 ${profile.detectedModels.length} 个模型`;}catch(e){profile.detectedModels=demo;profile.status='demo';profile.lastTest='展示示例目录 · 真实识别需 API/CORS/后端代理';}finally{this.modelDetecting=false;this.persist();this.showToast(profile.lastTest);}},
    inferModelCapabilities(name){const id=String(name||'').toLowerCase(),r=[];if(/vision|vl|multimodal/.test(id))r.push('视觉');if(/tts|speech|audio|voice/.test(id))r.push('语音');if(/image|draw|flux|sdxl/.test(id))r.push('生图');if(/video|wan/.test(id))r.push('视频');if(!r.length)r.push('对话');return r;},
    applyDetectedModel(profile,model){profile.chatModel=model.id;this.persist();this.showToast(`已将 ${model.id} 设为主对话模型`);},
    toggleCapability(key){this.settings.globalCapabilities[key]=!this.settings.globalCapabilities[key];this.showToast(this.settings.globalCapabilities[key]?'已启用前端配置占位':'已关闭');},
    validateRequestJson(){try{JSON.parse(this.settings.advanced.customRequestJson||'{}');this.showToast('JSON 格式正确');}catch(e){this.showToast('JSON 格式有误，请检查逗号和引号');}},
    exportData(){const b=new Blob([JSON.stringify(QinbanStore.exportData(),null,2)],{type:'application/json;charset=utf-8'}),u=URL.createObjectURL(b),a=document.createElement('a');a.href=u;a.download=`qinban-frontend-data-${new Date().toISOString().slice(0,10)}.json`;document.body.appendChild(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(u),0);this.showToast('已导出不含 API Key 的前端数据');},
    triggerImport(){this.$refs.importFile?.click();},
    async importData(event){const f=event.target.files?.[0];event.target.value='';if(!f)return;try{const data=JSON.parse(await f.text());QinbanStore.saveState(data);const state=QinbanStore.getState();this.companions=state.companions;this.groups=state.groups;this.moments=state.moments;this.apiProfiles=state.apiProfiles;this.favorites=state.favorites;this.settings=state.settings;this.showToast('本地数据已导入');}catch(e){this.showToast('导入失败：文件格式不正确');}},
    exportLogs(){const log={exportedAt:new Date().toISOString(),version:this.settings.version,frontendOnly:true,stats:{companions:this.companions.length,groups:this.groups.length,moments:this.moments.length,memories:this.memoryCount,messages:this.totalMessages},apiProfiles:this.apiProfiles.map(x=>({name:x.name,provider:x.provider,status:x.status,lastTest:x.lastTest}))};const b=new Blob([JSON.stringify(log,null,2)],{type:'application/json'}),u=URL.createObjectURL(b),a=document.createElement('a');a.href=u;a.download='qinban-frontend-log.json';document.body.appendChild(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(u),0);this.showToast('前端诊断日志已导出');},
    resetDemo(){if(!confirm('恢复演示数据会覆盖当前浏览器中的亲伴数据，确定继续吗？'))return;const s=QinbanStore.reset();this.companions=s.companions;this.groups=s.groups;this.moments=s.moments;this.apiProfiles=s.apiProfiles;this.favorites=s.favorites;this.settings=s.settings;this.selectedApiId=s.apiProfiles[0].id;this.activeChatRef=null;this.showToast('演示数据已恢复');},
    clearAllData(){if(!confirm('确定清空所有本地演示数据吗？此操作不可恢复。'))return;this.companions=[];this.groups=[];this.moments=[];this.favorites=[];this.persist();this.activeChatRef=null;this.showToast('本地业务数据已清空');},
    closeModal(){this.showGroupDialog=false;this.showProfileDialog=false;this.showMemoryDialog=false;this.showMomentDialog=false;this.utilityPanel='';},
    showToast(message){this.toastMessage=message;if(this.toastTimer)clearTimeout(this.toastTimer);this.toastTimer=setTimeout(()=>{this.toastMessage='';},2600);}
  }
};

