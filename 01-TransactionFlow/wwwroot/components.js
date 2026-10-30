const UIComponents = {
  renderInventoryDropdown(inventoryDict, targetSelect) {
    targetSelect.innerHTML = '<option value="">-- PILIH BARANG (INVENTORY) --</option>';
    if (!inventoryDict) return;
    Object.keys(inventoryDict).forEach((id) => {
      const inv = inventoryDict[id];
      const itemName = inv.ItemName || inv.itemName || inv.Item_Name || "Unknown";
      const stock = inv.Stock !== undefined ? inv.Stock : inv.stock !== undefined ? inv.stock : 0;
      const option = document.createElement("option");
      option.value = id;
      option.textContent = `${itemName} [STOK: ${stock}] (ID: ${id})`;
      targetSelect.appendChild(option);
    });
  },

  renderEmployeeOptions(employeesDict, targetSelect) {
    targetSelect.innerHTML = '<option value="">-- PILIH EMPLOYEE --</option>';
    if (!employeesDict) return;
    Object.keys(employeesDict).forEach((id) => {
      const emp = employeesDict[id];
      const name = emp.EmpName || emp.empName || emp.Name || "Unknown";
      const option = document.createElement("option");
      option.value = id;
      option.textContent = `[${id}] ${name}`;
      targetSelect.appendChild(option);
    });
  },

  renderCustomerOptions(customersDict, targetSelect) {
    targetSelect.innerHTML = '<option value="">-- PILIH CUSTOMER --</option>';
    if (!customersDict) return;
    Object.keys(customersDict).forEach((id) => {
      const cust = customersDict[id];
      const name = cust.CustName || cust.custName || cust.Name || "Unknown";
      const isMember = cust.IsMember || cust.isMember ? "(MEMBER)" : "";
      const option = document.createElement("option");
      option.value = id;
      option.textContent = `[${id}] ${name} ${isMember}`;
      targetSelect.appendChild(option);
    });
  },

  renderEmployeeTable(employeesDict, container, onDeleteClick) {
    container.innerHTML = "";
    if (!employeesDict || Object.keys(employeesDict).length === 0) {
      container.innerHTML = '<tr><td colspan="3" style="text-align:center; color: var(--text-muted);">TIDAK ADA DATA PEGAWAI</td></tr>';
      return;
    }
    Object.keys(employeesDict).forEach((id) => {
      const emp = employeesDict[id];
      const name = emp.EmpName || emp.empName || emp.Name || "-";
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td style="color: var(--accent-amber);">${id}</td>
        <td>${name}</td>
        <td><button class="btn-danger btn-del-emp" data-id="${id}">HAPUS</button></td>
      `;
      container.appendChild(tr);
    });

    container.querySelectorAll(".btn-del-emp").forEach((btn) => {
      btn.addEventListener("click", () => onDeleteClick(btn.dataset.id));
    });
  },

  renderCustomerTable(customersDict, container, onDeleteClick) {
    container.innerHTML = "";
    if (!customersDict || Object.keys(customersDict).length === 0) {
      container.innerHTML = '<tr><td colspan="4" style="text-align:center; color: var(--text-muted);">TIDAK ADA DATA PELANGGAN</td></tr>';
      return;
    }
    Object.keys(customersDict).forEach((id) => {
      const cust = customersDict[id];
      const name = cust.CustName || cust.custName || cust.Name || "-";
      const isMember = cust.IsMember || cust.isMember ? "YES" : "NO";
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td style="color: var(--accent-amber);">${id}</td>
        <td>${name}</td>
        <td>${isMember}</td>
        <td><button class="btn-danger btn-del-cust" data-id="${id}">HAPUS</button></td>
      `;
      container.appendChild(tr);
    });

    container.querySelectorAll(".btn-del-cust").forEach((btn) => {
      btn.addEventListener("click", () => onDeleteClick(btn.dataset.id));
    });
  },

  renderTransactionsTable(transactionsDict, container) {
    container.innerHTML = "";
    if (!transactionsDict || Object.keys(transactionsDict).length === 0) {
      container.innerHTML = '<tr><td colspan="7" style="text-align:center; color: var(--text-muted);">BELUM ADA RIWAYAT TRANSAKSI</td></tr>';
      return;
    }
    Object.keys(transactionsDict).forEach((id) => {
      const tx = transactionsDict[id];
      const transId = tx.TransID || tx.transId || id || "-";
      const invId = tx.ItemID || tx.itemId || tx.InventoryID || "-";
      const empId = tx.EmpID || tx.empId || "-";
      const custId = tx.CustID || tx.custId || "-";
      const amount = tx.Amount || tx.amount || 0;
      const timestamp = tx.Timestamp || tx.timestamp || "-";
      const desc = tx.Desc || tx.desc || "-";

      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td style="color: var(--accent-amber);">${transId}</td>
        <td>${invId}</td>
        <td>${empId}</td>
        <td>${custId}</td>
        <td>${amount}</td>
        <td>${timestamp}</td>
        <td>${desc}</td>
      `;
      container.appendChild(tr);
    });
  },

  renderInventoryCards(inventoryDict, container, searchQuery, onEditClick, onDeleteClick) {
    container.innerHTML = "";
    if (!inventoryDict || Object.keys(inventoryDict).length === 0) {
      container.innerHTML = '<div style="grid-column: 1/-1; text-align:center; color: var(--text-muted);">TIDAK ADA ITEM STOK</div>';
      return;
    }

    const query = searchQuery ? searchQuery.toLowerCase().trim() : "";
    let count = 0;

    Object.keys(inventoryDict).forEach((id) => {
      const inv = inventoryDict[id];
      const itemName = inv.ItemName || inv.itemName || "-";
      const stock = inv.Stock !== undefined ? inv.Stock : inv.stock !== undefined ? inv.stock : 0;
      const flag = inv.ItemFlag !== undefined ? inv.ItemFlag : inv.itemFlag !== undefined ? inv.itemFlag : "-";
      const itemExp = inv.ItemExp || inv.itemExp || inv.Date || "-";

      if (query && !id.toLowerCase().includes(query) && !itemName.toLowerCase().includes(query)) {
        return;
      }

      count++;
      const card = document.createElement("div");
      card.className = "inv-card";
      card.innerHTML = `
        <div>
          <div class="inv-card-header">
            <span class="inv-id">ID: ${id}</span>
            <span class="inv-flag">FLAG: ${flag}</span>
          </div>
          <div class="inv-name">${itemName}</div>
        </div>
        <div>
          <div class="inv-meta">STOK: <strong>${stock}</strong></div>
          <div class="inv-meta" style="font-size:0.65rem; color:var(--text-muted);">EXP/DATE: ${itemExp}</div>
          <div class="inv-actions">
            <button class="btn-subtle btn-edit-inv" data-id="${id}">EDIT NAMA</button>
            <button class="btn-danger btn-del-inv" data-id="${id}">HAPUS</button>
          </div>
        </div>
      `;
      container.appendChild(card);
    });

    if (count === 0) {
      container.innerHTML = `<div style="grid-column: 1/-1; text-align:center; color: var(--text-muted);">TIDAK ADA ITEM DENGAN KATA KUNCI "${searchQuery}"</div>`;
    }

    container.querySelectorAll(".btn-edit-inv").forEach((btn) => {
      btn.addEventListener("click", () => {
        const id = btn.dataset.id;
        onEditClick(id, inventoryDict[id]);
      });
    });

    container.querySelectorAll(".btn-del-inv").forEach((btn) => {
      btn.addEventListener("click", () => onDeleteClick(btn.dataset.id));
    });
  },

  showModal(title, htmlContent) {
    const modal = document.getElementById("bends-modal");
    const titleEl = document.getElementById("modal-title");
    const bodyEl = document.getElementById("modal-body");

    titleEl.textContent = `// ${title}`;
    bodyEl.innerHTML = htmlContent;
    modal.classList.remove("hidden");
  },

  closeModal() {
    const modal = document.getElementById("bends-modal");
    modal.classList.add("hidden");
    document.getElementById("modal-body").innerHTML = "";
  }
};