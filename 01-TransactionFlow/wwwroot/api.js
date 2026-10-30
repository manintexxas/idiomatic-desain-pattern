const API_BASE = "";

async function handleResponse(res) {
  if (!res.ok) {
    const errorText = await res.text();
    throw new Error(errorText || `HTTP Error: ${res.status}`);
  }
  if (res.status === 204) {
    return null;
  }
  return await res.json();
}

const ApiService = {
  async getInventory() {
    const res = await fetch(`${API_BASE}/inventory`);
    return handleResponse(res);
  },

  async createInventory(data) {
    const res = await fetch(`${API_BASE}/inventory`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    return handleResponse(res);
  },

  async updateInventory(id, data) {
    const res = await fetch(`${API_BASE}/inventory/${encodeURIComponent(id)}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    return handleResponse(res);
  },

  async deleteInventory(id) {
    const res = await fetch(`${API_BASE}/inventory/${encodeURIComponent(id)}`, {
      method: "DELETE"
    });
    return handleResponse(res);
  },

  async addStock(invId, data) {
    const res = await fetch(`${API_BASE}/inventory/${encodeURIComponent(invId)}/add`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    return handleResponse(res);
  },

  async reduceStock(invId, data) {
    const res = await fetch(`${API_BASE}/inventory/${encodeURIComponent(invId)}/reduce`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    return handleResponse(res);
  },

  async getEmployees() {
    const res = await fetch(`${API_BASE}/employee`);
    return handleResponse(res);
  },

  async createEmployee(data) {
    const res = await fetch(`${API_BASE}/employee`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    return handleResponse(res);
  },

  async deleteEmployee(id) {
    const res = await fetch(`${API_BASE}/employee/${encodeURIComponent(id)}`, {
      method: "DELETE"
    });
    return handleResponse(res);
  },

  async getCustomers() {
    const res = await fetch(`${API_BASE}/customer`);
    return handleResponse(res);
  },

  async createCustomer(data) {
    const res = await fetch(`${API_BASE}/customer`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data)
    });
    return handleResponse(res);
  },

  async deleteCustomer(id) {
    const res = await fetch(`${API_BASE}/customer/${encodeURIComponent(id)}`, {
      method: "DELETE"
    });
    return handleResponse(res);
  },

  async getTransactions() {
    const res = await fetch(`${API_BASE}/transactions`);
    return handleResponse(res);
  }
};