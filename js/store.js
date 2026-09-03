/**
 * 青伴 - 数据存储模块
 */
const Store = {
  KEYS: {
    COMPANIONS: 'qingban_companions',
    MESSAGES: 'qingban_messages',
    SETTINGS: 'qingban_settings',
    MEMORY: 'qingban_memory'
  },
  getCompanions() {
    const data = localStorage.getItem(this.KEYS.COMPANIONS);
    return data ? JSON.parse(data) : [];
  },
  saveCompanions(companions) {
    localStorage.setItem(this.KEYS.COMPANIONS, JSON.stringify(companions));
  },
  addCompanion(companion) {
    const companions = this.getCompanions();
    companion.id = Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    companion.createdAt = Date.now();
    companions.push(companion);
    this.saveCompanions(companions);
    return companion;
  },
  updateCompanion(id, updates) {
    const companions = this.getCompanions();
    const index = companions.findIndex(c => c.id === id);
    if (index !== -1) {
      companions[index] = { ...companions[index], ...updates };
      this.saveCompanions(companions);
    }
  },
  deleteCompanion(id) {
    const companions = this.getCompanions().filter(c => c.id !== id);
    this.saveCompanions(companions);
    const messages = this.getMessages();
    delete messages[id];
    this.saveMessages(messages);
  },
  getMessages() {
    const data = localStorage.getItem(this.KEYS.MESSAGES);
    return data ? JSON.parse(data) : {};
  },
  saveMessages(messages) {
    localStorage.setItem(this.KEYS.MESSAGES, JSON.stringify(messages));
  },
  getChatMessages(companionId) {
    const messages = this.getMessages();
    return messages[companionId] || [];
  },
  addMessage(companionId, message) {
    const messages = this.getMessages();
    if (!messages[companionId]) { messages[companionId] = []; }
    message.id = Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    message.timestamp = Date.now();
    messages[companionId].push(message);
    this.saveMessages(messages);
    return message;
  },
  getSettings() {
    const data = localStorage.getItem(this.KEYS.SETTINGS);
    return data ? JSON.parse(data) : { nickname: '', signature: '', autoMsgNotify: true };
  },
  saveSettings(settings) {
    localStorage.setItem(this.KEYS.SETTINGS, JSON.stringify(settings));
  },
  clearAll() {
    Object.values(this.KEYS).forEach(key => { localStorage.removeItem(key); });
  },
  exportData() {
    return {
      companions: this.getCompanions(),
      messages: this.getMessages(),
      settings: this.getSettings(),
      exportTime: new Date().toISOString()
    };
  }
};