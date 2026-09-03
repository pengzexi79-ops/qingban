/**
 * 青伴 - 数据存储模块
 * 使用 localStorage 持久化数据
 */

const Store = {
  // 存储键名
  KEYS: {
    COMPANIONS: 'qingban_companions',
    MESSAGES: 'qingban_messages',
    SETTINGS: 'qingban_settings',
    MEMORY: 'qingban_memory'
  },

  // 获取所有AI伴侣
  getCompanions() {
    const data = localStorage.getItem(this.KEYS.COMPANIONS);
    return data ? JSON.parse(data) : [];
  },

  // 保存AI伴侣列表
  saveCompanions(companions) {
    localStorage.setItem(this.KEYS.COMPANIONS, JSON.stringify(companions));
  },

  // 添加AI伴侣
  addCompanion(companion) {
    const companions = this.getCompanions();
    companion.id = Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    companion.createdAt = Date.now();
    companions.push(companion);
    this.saveCompanions(companions);
    return companion;
  },

  // 更新AI伴侣
  updateCompanion(id, updates) {
    const companions = this.getCompanions();
    const index = companions.findIndex(c => c.id === id);
    if (index !== -1) {
      companions[index] = { ...companions[index], ...updates };
      this.saveCompanions(companions);
    }
  },

  // 删除AI伴侣
  deleteCompanion(id) {
    const companions = this.getCompanions().filter(c => c.id !== id);
    this.saveCompanions(companions);
    // 同时删除相关消息
    const messages = this.getMessages();
    delete messages[id];
    this.saveMessages(messages);
  },

  // 获取所有消息
  getMessages() {
    const data = localStorage.getItem(this.KEYS.MESSAGES);
    return data ? JSON.parse(data) : {};
  },

  // 保存消息
  saveMessages(messages) {
    localStorage.setItem(this.KEYS.MESSAGES, JSON.stringify(messages));
  },

  // 获取与某个伴侣的聊天记录
  getChatMessages(companionId) {
    const messages = this.getMessages();
    return messages[companionId] || [];
  },

  // 添加消息
  addMessage(companionId, message) {
    const messages = this.getMessages();
    if (!messages[companionId]) {
      messages[companionId] = [];
    }
    message.id = Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    message.timestamp = Date.now();
    messages[companionId].push(message);
    this.saveMessages(messages);
    return message;
  },

  // 获取用户设置
  getSettings() {
    const data = localStorage.getItem(this.KEYS.SETTINGS);
    return data ? JSON.parse(data) : {
      nickname: '',
      signature: '',
      autoMsgNotify: true
    };
  },

  // 保存用户设置
  saveSettings(settings) {
    localStorage.setItem(this.KEYS.SETTINGS, JSON.stringify(settings));
  },

  // 记忆库相关
  getMemory() {
    const data = localStorage.getItem(this.KEYS.MEMORY);
    return data ? JSON.parse(data) : {};
  },

  saveMemory(memory) {
    localStorage.setItem(this.KEYS.MEMORY, JSON.stringify(memory));
  },

  // 添加记忆条目
  addMemoryEntry(companionId, entry) {
    const memory = this.getMemory();
    if (!memory[companionId]) {
      memory[companionId] = [];
    }
    entry.id = Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    entry.timestamp = Date.now();
    memory[companionId].push(entry);
    this.saveMemory(memory);
    return entry;
  },

  // 清除所有数据
  clearAll() {
    Object.values(this.KEYS).forEach(key => {
      localStorage.removeItem(key);
    });
  },

  // 导出数据
  exportData() {
    return {
      companions: this.getCompanions(),
      messages: this.getMessages(),
      settings: this.getSettings(),
      memory: this.getMemory(),
      exportTime: new Date().toISOString()
    };
  }
};
