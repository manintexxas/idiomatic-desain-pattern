document.addEventListener("DOMContentLoaded", () => {
  let state = {
    inventory: {},
    employees: {},
    customers: {},
    transactions: {}
  };

  initNavigation();
  initModalListeners();
  initFormListeners();
  initSearchListener();
  
  loadAllData();

  function initNavigation() {
    const navButtons = document.querySelectorAll(".nav-btn");
    const tabContents = document.querySelectorAll(".tab-content");

    navButtons.forEach((btn) => {
      btn.addEventListener("click", () => {
        const targetTab = btn.dataset.tab;

        navButtons.forEach((b) => b.classList.remove("active"));
        tabContents.forEach((tc) => tc.classList.remove("active"));

        btn.classList.add("active");
        document.getElementById(targetTab).classList.add("active");
      });
    });

    document.getElementById("btn-refresh-tx").addEventListener("click", fetchTransactions);
    document.getElementById("btn-open-create-inv").addEventListener("click", openCreateInventoryModal);
  }

  function initModalListeners() {
    document.getElementById("modal-close-btn").addEventListener("click", () => {
      UIComponents.closeModal();
    });

    window.addEventListener("click", (e) => {
      const modal = document.getElementById("bends-modal");
      if (e.target === modal) {
        UIComponents.closeModal();
      }
    });
  }

  async function loadAllData() {
    await Promise.all([
      fetchInventory(),
      fetchEmployees(),
      fetchCustomers(),
      fetchTransactions()
    ]);
  }

  async function fetchInventory() {
    try {
      state.inventory = await ApiService.getInventory() || {};
      renderInventoryUI();
      renderInventoryDropdowns();
    } catch (err) {
      console.error("Gagal memuat inventory:", err);
    }
  }

  async function fetchEmployees() {
    try {
      state.employees = await ApiService.getEmployees() || {};
      renderEmployeesUI();
    } catch (err) {
      console.error("Gagal memuat employee:", err);
    }
  }

  async function fetchCustomers() {
    try {
      state.customers = await ApiService.getCustomers() || {};
      renderCustomersUI();
    } catch (err) {
      console.error("Gagal memuat customer:", err);
    }
  }

  async function fetchTransactions() {
    try {
      state.transactions = await ApiService.getTransactions() || {};
      const container = document.getElementById("transactions-table-body");
      UIComponents.renderTransactionsTable(state.transactions, container);
    } catch (err) {
      console.error("Gagal memuat transaksi:", err);
    }
  }

  function renderInventoryUI() {
    const searchInput = document.getElementById("inventory-search-input");
    const container = document.getElementById("inventory-cards-grid");
    
    UIComponents.renderInventoryCards(
      state.inventory,
      container,
      searchInput.value,
      handleEditInventory,
      handleDeleteInventory
    );
  }

  function renderInventoryDropdowns() {
    const reduceDropdown = document.getElementById("reduce-inv-dropdown");
    const addDropdown = document.getElementById("add-inv-dropdown");

    UIComponents.renderInventoryDropdown(state.inventory, reduceDropdown);
    UIComponents.renderInventoryDropdown(state.inventory, addDropdown);
  }

  function renderEmployeesUI() {
    const empTable = document.getElementById("employee-table-body");
    const reduceEmpSelect = document.getElementById("reduce-emp-id");
    const addReceiverSelect = document.getElementById("add-receiver-id");

    UIComponents.renderEmployeeTable(state.employees, empTable, handleDeleteEmployee);
    UIComponents.renderEmployeeOptions(state.employees, reduceEmpSelect);
    UIComponents.renderEmployeeOptions(state.employees, addReceiverSelect);
  }

  function renderCustomersUI() {
    const custTable = document.getElementById("customer-table-body");
    const reduceCustSelect = document.getElementById("reduce-cust-id");

    UIComponents.renderCustomerTable(state.customers, custTable, handleDeleteCustomer);
    UIComponents.renderCustomerOptions(state.customers, reduceCustSelect);
  }

  function initSearchListener() {
    const searchInput = document.getElementById("inventory-search-input");
    searchInput.addEventListener("input", () => {
      renderInventoryUI();
    });
  }

  function initFormListeners() {
    document.getElementById("form-reduce-stock").addEventListener("submit", async (e) => {
      e.preventDefault();
      const invId = document.getElementById("reduce-inv-dropdown").value;
      const empId = document.getElementById("reduce-emp-id").value;
      const custId = document.getElementById("reduce-cust-id").value;
      const amount = parseInt(document.getElementById("reduce-amount").value, 10);
      const desc = document.getElementById("reduce-desc").value;

      try {
        await ApiService.reduceStock(invId, { 
          empId: empId, 
          custId: custId, 
          amount: amount, 
          desc: desc 
        });
        e.target.reset();
        await Promise.all([fetchInventory(), fetchTransactions()]);
        UIComponents.showModal("SUCCESS", "<p style='color: var(--accent-amber);'>TRANSAKSI OUTBOUND BERHASIL DIEKSEKUSI.</p>");
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });

    document.getElementById("form-add-stock").addEventListener("submit", async (e) => {
      e.preventDefault();
      const invId = document.getElementById("add-inv-dropdown").value;
      const receiverId = document.getElementById("add-receiver-id").value;
      const amount = parseInt(document.getElementById("add-amount").value, 10);
      const desc = document.getElementById("add-desc").value;

      try {
        await ApiService.addStock(invId, { 
          receiverId: receiverId, 
          amount: amount, 
          desc: desc 
        });
        e.target.reset();
        await Promise.all([fetchInventory(), fetchTransactions()]);
        UIComponents.showModal("SUCCESS", "<p style='color: var(--accent-amber);'>TRANSAKSI INBOUND BERHASIL DIEKSEKUSI.</p>");
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });

    document.getElementById("form-create-employee").addEventListener("submit", async (e) => {
      e.preventDefault();
      const empName = document.getElementById("emp-name").value;
      try {
        await ApiService.createEmployee({ empName });
        e.target.reset();
        await fetchEmployees();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });

    document.getElementById("form-create-customer").addEventListener("submit", async (e) => {
      e.preventDefault();
      const custName = document.getElementById("cust-name").value;
      const isMember = document.getElementById("cust-member").checked;
      try {
        await ApiService.createCustomer({ custName, isMember });
        e.target.reset();
        await fetchCustomers();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });
  }

  function showConfirmModal(title, message, onConfirm) {
    const html = `
      <div style="display: flex; flex-direction: column; gap: 1rem;">
        <p style="color: var(--text-main); font-size: 0.8rem;">${message}</p>
        <div style="display: flex; gap: 1rem; justify-content: flex-end; margin-top: 0.5rem;">
          <button id="modal-btn-cancel" class="btn-subtle">BATAL</button>
          <button id="modal-btn-yes" class="btn-danger">YA, HAPUS</button>
        </div>
      </div>
    `;
    UIComponents.showModal(title, html);

    document.getElementById("modal-btn-cancel").addEventListener("click", () => {
      UIComponents.closeModal();
    });

    document.getElementById("modal-btn-yes").addEventListener("click", async () => {
      UIComponents.closeModal();
      await onConfirm();
    });
  }

  async function handleDeleteEmployee(id) {
    showConfirmModal("KONFIRMASI HAPUS", `Hapus data pegawai [${id}] dari sistem?`, async () => {
      try {
        await ApiService.deleteEmployee(id);
        await fetchEmployees();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });
  }

  async function handleDeleteCustomer(id) {
    showConfirmModal("KONFIRMASI HAPUS", `Hapus data pelanggan [${id}] dari sistem?`, async () => {
      try {
        await ApiService.deleteCustomer(id);
        await fetchCustomers();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });
  }

  async function handleDeleteInventory(id) {
    showConfirmModal("KONFIRMASI HAPUS", `Hapus item inventory [${id}] dari sistem?`, async () => {
      try {
        await ApiService.deleteInventory(id);
        await fetchInventory();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });
  }

  function handleEditInventory(id, invData) {
    const itemName = invData.ItemName || invData.itemName || "";
    const stock = invData.Stock !== undefined ? invData.Stock : invData.stock || 0;
    const flag = invData.ItemFlag !== undefined ? invData.ItemFlag : invData.itemFlag || 0;

    const html = `
      <form id="form-edit-inv" class="retro-form">
        <div class="field-group">
          <label>ID ITEM (LOCKED)</label>
          <input type="text" value="${id}" disabled readonly>
        </div>
        <div class="field-group">
          <label>NAMA ITEM BARU</label>
          <input type="text" id="edit-inv-name" value="${itemName}" required>
        </div>
        <button type="submit" class="btn-industrial">SIMPAN PERUBAHAN NAMA</button>
      </form>
    `;

    UIComponents.showModal(`EDIT ITEM [${id}]`, html);

    document.getElementById("form-edit-inv").addEventListener("submit", async (e) => {
      e.preventDefault();
      const newName = document.getElementById("edit-inv-name").value;

      try {
        await ApiService.updateInventory(id, {
          itemName: newName,
          stock: stock,
          flag: flag
        });
        UIComponents.closeModal();
        await fetchInventory();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });
  }

  function openCreateInventoryModal() {
    const html = `
      <form id="form-create-inv" class="retro-form">
        <div class="field-group">
          <label>TANGGAL (YYYY-MM-DD)</label>
          <input type="text" id="create-inv-date" placeholder="2026-10-30" required>
        </div>
        <div class="field-group">
          <label>NAMA ITEM</label>
          <input type="text" id="create-inv-name" placeholder="Nama barang..." required>
        </div>
        <div class="field-group">
          <label>STOK AWAL</label>
          <input type="number" id="create-inv-stock" min="0" value="0" required>
        </div>
        <div class="field-group">
          <label>FLAG ITEM (ANGKA)</label>
          <input type="number" id="create-inv-flag" value="1" required>
        </div>
        <button type="submit" class="btn-industrial">+ BUAT ITEM BARU</button>
      </form>
    `;

    UIComponents.showModal("BUAT ITEM INVENTORY BARU", html);

    document.getElementById("form-create-inv").addEventListener("submit", async (e) => {
      e.preventDefault();
      const rawDate = document.getElementById("create-inv-date").value;
      const date = rawDate.includes("T") ? rawDate.replace("T", " ").replace("Z", "") : `${rawDate} 00:00:00`;

      const itemName = document.getElementById("create-inv-name").value;
      const stock = parseInt(document.getElementById("create-inv-stock").value, 10);
      const flag = parseInt(document.getElementById("create-inv-flag").value, 10);

      try {
        await ApiService.createInventory({ date, itemName, stock, flag });
        UIComponents.closeModal();
        await fetchInventory();
      } catch (err) {
        UIComponents.showModal("BACKEND ERROR", `<p style='color: var(--accent-alert);'>${err.message}</p>`);
      }
    });
  }
});