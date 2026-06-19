package web

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PostgreSQL Client</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        
        .login-overlay {
            position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center;
            z-index: 1000;
        }
        .login-box {
            background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);
            width: 400px;
        }
        .login-box h2 { margin-bottom: 20px; color: #333; }
        .login-box input {
            width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 4px;
            font-size: 14px; margin-bottom: 15px;
        }
        .login-box button {
            width: 100%; padding: 12px; background: #4a90d9; color: white; border: none;
            border-radius: 4px; font-size: 16px; cursor: pointer;
        }
        .login-box button:hover { background: #357abd; }
        .login-error { color: #e74c3c; margin-bottom: 10px; display: none; }

        .header {
            background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .header h1 { color: #333; font-size: 24px; }
        .header-info { color: #666; margin-top: 5px; }

        .sidebar {
            position: fixed; left: 0; top: 0; width: 280px; height: 100%;
            background: #2c3e50; color: white; overflow-y: auto; z-index: 100;
        }
        .sidebar-header { padding: 20px; background: #1a252f; }
        .sidebar-header h2 { font-size: 18px; margin-bottom: 5px; }
        .sidebar-header p { font-size: 12px; color: #95a5a6; }
        .sidebar-nav { padding: 10px 0; }
        .nav-item {
            padding: 12px 20px; cursor: pointer; transition: background 0.2s;
            border-left: 3px solid transparent;
        }
        .nav-item:hover { background: #34495e; }
        .nav-item.active { background: #3498db; border-left-color: #fff; }
        .nav-section { padding: 10px 20px; color: #7f8c8d; font-size: 12px; text-transform: uppercase; }

        .main-content { margin-left: 280px; padding: 20px; }

        .card {
            background: white; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .card-header {
            padding: 15px 20px; border-bottom: 1px solid #eee;
            display: flex; justify-content: space-between; align-items: center;
        }
        .card-header h3 { font-size: 16px; color: #333; }
        .card-body { padding: 20px; }

        .btn {
            padding: 8px 16px; border-radius: 4px; border: none; cursor: pointer;
            font-size: 14px; transition: all 0.2s;
        }
        .btn-primary { background: #4a90d9; color: white; }
        .btn-primary:hover { background: #357abd; }
        .btn-success { background: #27ae60; color: white; }
        .btn-success:hover { background: #219a52; }
        .btn-danger { background: #e74c3c; color: white; }
        .btn-danger:hover { background: #c0392b; }
        .btn-secondary { background: #95a5a6; color: white; }
        .btn-secondary:hover { background: #7f8c8d; }
        .btn-sm { padding: 5px 10px; font-size: 12px; }

        .table-container { overflow-x: auto; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px 15px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: 600; color: #333; }
        tr:hover { background: #f8f9fa; }
        td { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        td:hover { overflow: visible; white-space: normal; word-break: break-all; }

        .pagination {
            display: flex; align-items: center; justify-content: space-between;
            padding: 15px 0;
        }
        .pagination-info { color: #666; font-size: 14px; }
        .pagination-btns { display: flex; gap: 5px; }
        .pagination-btns button { min-width: 36px; }
        .pagination-btns button.active { background: #4a90d9; }

        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: 500; color: #333; }
        .form-control {
            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px;
            font-size: 14px;
        }
        .form-control:focus { outline: none; border-color: #4a90d9; }

        .modal {
            display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;
        }
        .modal.active { display: flex; }
        .modal-content {
            background: white; border-radius: 8px; width: 90%; max-width: 600px;
            max-height: 80vh; overflow-y: auto;
        }
        .modal-header {
            padding: 20px; border-bottom: 1px solid #eee;
            display: flex; justify-content: space-between; align-items: center;
        }
        .modal-header h3 { font-size: 18px; }
        .modal-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #999; }
        .modal-body { padding: 20px; }
        .modal-footer { padding: 15px 20px; border-top: 1px solid #eee; text-align: right; }

        .structure-table th { background: #ecf0f1; }
        .nullable-yes { color: #27ae60; }
        .nullable-no { color: #e74c3c; }

        .upload-zone {
            border: 2px dashed #ddd; border-radius: 8px; padding: 40px; text-align: center;
            transition: all 0.3s; cursor: pointer;
        }
        .upload-zone:hover, .upload-zone.dragover {
            border-color: #4a90d9; background: #f8f9fa;
        }
        .upload-zone p { color: #666; margin-top: 10px; }

        .alert { padding: 12px 15px; border-radius: 4px; margin-bottom: 15px; }
        .alert-success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .alert-error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }

        .hidden { display: none; }

        .table-list { list-style: none; }
        .table-list-item {
            padding: 10px 20px; cursor: pointer; transition: background 0.2s;
            display: flex; justify-content: space-between; align-items: center;
        }
        .table-list-item:hover { background: #34495e; }
        .table-list-item.active { background: #3498db; }
        .table-list-item .count { font-size: 12px; color: #95a5a6; }

        .empty-state { text-align: center; padding: 40px; color: #999; }

        .column-compare { margin-top: 10px; }
        .column-compare-header { display: flex; gap: 20px; margin-bottom: 8px; font-weight: 500; color: #333; font-size: 13px; }
        .column-compare-row { display: flex; gap: 20px; padding: 4px 0; font-size: 13px; border-bottom: 1px solid #f0f0f0; }
        .column-compare-row .csv-col { flex: 1; }
        .column-compare-row .tbl-col { flex: 1; }
        .column-compare-row .status { width: 60px; text-align: center; }
        .column-match { color: #27ae60; }
        .column-mismatch { color: #e74c3c; }
        .column-compare-summary { margin-top: 10px; padding: 8px 12px; border-radius: 4px; font-size: 13px; }
        .column-compare-summary.ok { background: #d4edda; color: #155724; }
        .column-compare-summary.fail { background: #f8d7da; color: #721c24; }

        .sidebar-actions { padding: 10px 20px; border-top: 1px solid #34495e; }
        .sidebar-actions .btn { width: 100%; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div id="loginOverlay" class="login-overlay">
        <div class="login-box">
            <h2>PostgreSQL Client</h2>
            <div id="loginError" class="login-error">Invalid password</div>
            <input type="password" id="loginPassword" placeholder="Enter password" autofocus>
            <button onclick="login()">Login</button>
        </div>
    </div>

    <div class="sidebar">
        <div class="sidebar-header">
            <h2>PostgreSQL Client</h2>
            <p>Database Management</p>
        </div>
        <div class="nav-section">Tables</div>
        <div class="sidebar-nav" id="tableList"></div>
        <div class="sidebar-actions">
            <button class="btn btn-success btn-sm" onclick="showCreateTableModal()">+ Create Table from CSV</button>
        </div>
    </div>

    <div class="main-content">
        <div class="header">
            <h1>PostgreSQL Web Client</h1>
            <div class="header-info" id="headerInfo">Select a table to get started</div>
        </div>

        <div id="dashboard" class="card">
            <div class="card-header">
                <h3>Dashboard</h3>
            </div>
            <div class="card-body">
                <div id="dashboardContent" class="empty-state">
                    <p>Select a table from the sidebar to view its data</p>
                </div>
            </div>
        </div>

        <div id="tableDataCard" class="card hidden">
            <div class="card-header">
                <h3 id="tableName">Table Data</h3>
                <div>
                    <button class="btn btn-success btn-sm" onclick="showInsertModal()">+ Insert Row</button>
                    <button class="btn btn-secondary btn-sm" onclick="showStructure()">Structure</button>
                    <button class="btn btn-primary btn-sm" onclick="showImportModal()">Import CSV</button>
                </div>
            </div>
            <div class="card-body">
                <div id="alertContainer"></div>
                <div class="table-container">
                    <table id="dataTable">
                        <thead id="tableHead"></thead>
                        <tbody id="tableBody"></tbody>
                    </table>
                </div>
                <div id="pagination" class="pagination"></div>
            </div>
        </div>

        <div id="structureCard" class="card hidden">
            <div class="card-header">
                <h3 id="structureTableName">Table Structure</h3>
                <button class="btn btn-secondary btn-sm" onclick="showDataTable()">Back to Data</button>
            </div>
            <div class="card-body">
                <div class="table-container">
                    <table id="structureTable">
                        <thead>
                            <tr>
                                <th>Column Name</th>
                                <th>Data Type</th>
                                <th>Nullable</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody id="structureBody"></tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <div id="insertModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3>Insert New Row</h3>
                <button class="modal-close" onclick="closeModal('insertModal')">&times;</button>
            </div>
            <div class="modal-body" id="insertFormBody"></div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="closeModal('insertModal')">Cancel</button>
                <button class="btn btn-success" onclick="insertRow()">Insert</button>
            </div>
        </div>
    </div>

    <div id="editModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3>Edit Row</h3>
                <button class="modal-close" onclick="closeModal('editModal')">&times;</button>
            </div>
            <div class="modal-body" id="editFormBody"></div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="closeModal('editModal')">Cancel</button>
                <button class="btn btn-primary" onclick="updateRow()">Save Changes</button>
            </div>
        </div>
    </div>

    <div id="importModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3 id="importModalTitle">Import CSV</h3>
                <button class="modal-close" onclick="closeModal('importModal')">&times;</button>
            </div>
            <div class="modal-body">
                <input type="hidden" id="importMode" value="import">
                <div id="importTableNameGroup" class="form-group hidden">
                    <label>Table Name</label>
                    <input type="text" id="importTableName" class="form-control" placeholder="Table name (defaults to filename)">
                </div>
                <div class="upload-zone" id="uploadZone">
                    <p>Drag & drop CSV file here or click to select</p>
                    <input type="file" id="csvFile" accept=".csv" style="display:none">
                </div>
                <div id="columnCompare" class="hidden column-compare">
                    <div class="column-compare-header">
                        <span style="flex:1;">CSV Columns</span>
                        <span style="flex:1;">Table Columns</span>
                        <span style="width:60px;text-align:center;">Status</span>
                    </div>
                    <div id="columnCompareBody"></div>
                    <div id="columnCompareSummary"></div>
                </div>
                <div id="columnEditor" class="hidden" style="margin-top:15px;">
                    <label style="font-weight:500;color:#333;margin-bottom:8px;display:block;">Column Names (editable)</label>
                    <div id="columnInputs"></div>
                </div>
                <div id="uploadProgress" class="hidden" style="margin-top:15px;">
                    <div style="background:#eee;border-radius:4px;height:20px;overflow:hidden;">
                        <div id="progressBar" style="background:#4a90d9;height:100%;width:0%;transition:width 0.3s;"></div>
                    </div>
                    <p id="progressText" style="margin-top:5px;color:#666;font-size:14px;">Uploading...</p>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="closeModal('importModal')">Cancel</button>
                <button class="btn btn-success" id="importBtn" onclick="importCSV()" disabled>Import</button>
            </div>
        </div>
    </div>

    <script>
        let currentTable = '';
        let currentPage = 1;
        let totalPages = 1;
        let currentColumns = [];
        let editOldData = null;
        let currentSort = '';
        let currentSortDir = 'asc';

        async function login() {
            const password = document.getElementById('loginPassword').value;
            try {
                const resp = await fetch('/api/login', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({password})
                });
                const data = await resp.json();
                if (data.success) {
                    document.getElementById('loginOverlay').style.display = 'none';
                    loadTables();
                } else {
                    document.getElementById('loginError').style.display = 'block';
                }
            } catch (e) {
                document.getElementById('loginError').textContent = 'Connection error';
                document.getElementById('loginError').style.display = 'block';
            }
        }

        document.getElementById('loginPassword').addEventListener('keypress', e => {
            if (e.key === 'Enter') login();
        });

        window.addEventListener('load', async () => {
            try {
                const resp = await fetch('/api/tables');
                if (resp.ok) {
                    document.getElementById('loginOverlay').style.display = 'none';
                    const data = await resp.json();
                    if (data.success) {
                        renderTableList(data.data);
                    }
                }
            } catch (e) {}
        });

        async function loadTables() {
            try {
                const resp = await fetch('/api/tables');
                const data = await resp.json();
                if (data.success) {
                    renderTableList(data.data);
                }
            } catch (e) {
                console.error('Failed to load tables:', e);
            }
        }

        function renderTableList(tables) {
            const container = document.getElementById('tableList');
            if (!tables || tables.length === 0) {
                container.innerHTML = '<div style="padding:10px 20px;color:#95a5a6;">No tables found</div>';
                return;
            }
            container.innerHTML = tables.map(t => 
                '<div class="table-list-item" onclick="selectTable(\'' + t.name + '\')">' +
                '<span>' + t.name + '</span>' +
                '<span class="count">' + t.rowCount + ' rows</span>' +
                '</div>'
            ).join('');
        }

        function selectTable(name) {
            currentTable = name;
            currentPage = 1;
            currentSort = '';
            currentSortDir = 'asc';
            document.querySelectorAll('.table-list-item').forEach(el => el.classList.remove('active'));
            event.currentTarget.classList.add('active');
            
            document.getElementById('dashboard').classList.add('hidden');
            document.getElementById('structureCard').classList.add('hidden');
            document.getElementById('tableDataCard').classList.remove('hidden');
            document.getElementById('tableName').textContent = name;
            document.getElementById('headerInfo').textContent = 'Viewing: ' + name;
            
            loadTableData();
        }

        async function loadTableData() {
            try {
                let url = '/api/table/' + encodeURIComponent(currentTable) + '?page=' + currentPage + '&pageSize=20';
                if (currentSort) {
                    url += '&sort=' + encodeURIComponent(currentSort) + '&order=' + currentSortDir;
                }
                const resp = await fetch(url);
                const data = await resp.json();
                if (data.success) {
                    renderTable(data);
                } else {
                    showAlert('error', data.error);
                }
            } catch (e) {
                showAlert('error', 'Failed to load data');
            }
        }

        function renderTable(data) {
            currentColumns = data.columns;
            totalPages = data.totalPage;
            
            document.getElementById('tableHead').innerHTML = '<tr>' + 
                data.columns.map(c => {
                    let arrow = '';
                    if (currentSort === c) {
                        arrow = currentSortDir === 'asc' ? ' ▲' : ' ▼';
                    }
                    return '<th style="cursor:pointer;user-select:none;" onclick="toggleSort(\'' + escapeHtml(c).replace(/'/g, "\\'") + '\')">' + escapeHtml(c) + arrow + '</th>';
                }).join('') +
                '<th>Actions</th></tr>';
            
            document.getElementById('tableBody').innerHTML = data.rows.map((row, idx) => {
                const rowJson = JSON.stringify(row).replace(/"/g, '&quot;');
                return '<tr>' +
                    data.columns.map(c => '<td title="' + escapeHtml(String(row[c] || '')) + '">' + escapeHtml(String(row[c] || 'NULL')) + '</td>').join('') +
                    '<td>' +
                        '<button class="btn btn-primary btn-sm" onclick="editRow(' + rowJson + ')">Edit</button> ' +
                        '<button class="btn btn-danger btn-sm" onclick="deleteRow(' + rowJson + ')">Delete</button>' +
                    '</td></tr>';
            }).join('');
            
            renderPagination(data.page, data.total, data.pageSize);
        }

        function renderPagination(page, total, pageSize) {
            const totalPages = Math.ceil(total / pageSize);
            let html = '<div class="pagination-info">Showing ' + ((page-1)*pageSize+1) + '-' + Math.min(page*pageSize, total) + ' of ' + total + ' rows</div>';
            html += '<div class="pagination-btns">';
            html += '<button class="btn btn-secondary btn-sm" onclick="goToPage(1)" ' + (page<=1?'disabled':'') + '>&laquo;</button>';
            html += '<button class="btn btn-secondary btn-sm" onclick="goToPage(' + (page-1) + ')" ' + (page<=1?'disabled':'') + '>&lsaquo;</button>';
            
            let start = Math.max(1, page - 2);
            let end = Math.min(totalPages, page + 2);
            for (let i = start; i <= end; i++) {
                html += '<button class="btn btn-sm ' + (i===page?'btn-primary':'btn-secondary') + '" onclick="goToPage(' + i + ')">' + i + '</button>';
            }
            
            html += '<button class="btn btn-secondary btn-sm" onclick="goToPage(' + (page+1) + ')" ' + (page>=totalPages?'disabled':'') + '>&rsaquo;</button>';
            html += '<button class="btn btn-secondary btn-sm" onclick="goToPage(' + totalPages + ')" ' + (page>=totalPages?'disabled':'') + '>&raquo;</button>';
            html += '</div>';
            
            document.getElementById('pagination').innerHTML = html;
        }

        function goToPage(page) {
            if (page < 1 || page > totalPages) return;
            currentPage = page;
            loadTableData();
        }

        function toggleSort(column) {
            if (currentSort === column) {
                currentSortDir = currentSortDir === 'asc' ? 'desc' : 'asc';
            } else {
                currentSort = column;
                currentSortDir = 'asc';
            }
            currentPage = 1;
            loadTableData();
        }

        async function showStructure() {
            try {
                const resp = await fetch('/api/describe/' + encodeURIComponent(currentTable));
                const data = await resp.json();
                if (data.success) {
                    document.getElementById('tableDataCard').classList.add('hidden');
                    document.getElementById('structureCard').classList.remove('hidden');
                    document.getElementById('structureTableName').textContent = currentTable + ' Structure';
                    
                    document.getElementById('structureBody').innerHTML = data.data.map(c =>
                        '<tr>' +
                        '<td>' + escapeHtml(c.name) + '</td>' +
                        '<td>' + escapeHtml(c.dataType) + '</td>' +
                        '<td class="' + (c.isNullable ? 'nullable-yes' : 'nullable-no') + '">' + (c.isNullable ? 'YES' : 'NO') + '</td>' +
                        '<td>' + escapeHtml(c.defaultValue || 'NULL') + '</td>' +
                        '</tr>'
                    ).join('');
                }
            } catch (e) {
                showAlert('error', 'Failed to load structure');
            }
        }

        function showDataTable() {
            document.getElementById('structureCard').classList.add('hidden');
            document.getElementById('tableDataCard').classList.remove('hidden');
        }

        function showInsertModal() {
            document.getElementById('insertFormBody').innerHTML = currentColumns.map(c =>
                '<div class="form-group"><label>' + escapeHtml(c) + '</label>' +
                '<input type="text" class="form-control" data-column="' + escapeHtml(c) + '" placeholder="Enter ' + escapeHtml(c) + '"></div>'
            ).join('');
            document.getElementById('insertModal').classList.add('active');
        }

        async function insertRow() {
            const row = {};
            document.querySelectorAll('#insertFormBody .form-control').forEach(input => {
                row[input.dataset.column] = input.value || null;
            });
            
            try {
                const resp = await fetch('/api/row/' + encodeURIComponent(currentTable) + '/insert', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(row)
                });
                const data = await resp.json();
                if (data.success) {
                    closeModal('insertModal');
                    showAlert('success', 'Row inserted successfully');
                    loadTableData();
                    loadTables();
                } else {
                    showAlert('error', data.error);
                }
            } catch (e) {
                showAlert('error', 'Failed to insert row');
            }
        }

        function editRow(row) {
            editOldData = row;
            document.getElementById('editFormBody').innerHTML = currentColumns.map(c =>
                '<div class="form-group"><label>' + escapeHtml(c) + '</label>' +
                '<input type="text" class="form-control" data-column="' + escapeHtml(c) + '" value="' + escapeHtml(String(row[c] || '')) + '"></div>'
            ).join('');
            document.getElementById('editModal').classList.add('active');
        }

        async function updateRow() {
            const newValues = {};
            document.querySelectorAll('#editFormBody .form-control').forEach(input => {
                newValues[input.dataset.column] = input.value || null;
            });
            
            try {
                const resp = await fetch('/api/row/' + encodeURIComponent(currentTable) + '/update/x', {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({newValues, oldValues: editOldData})
                });
                const data = await resp.json();
                if (data.success) {
                    closeModal('editModal');
                    showAlert('success', 'Row updated successfully');
                    loadTableData();
                } else {
                    showAlert('error', data.error);
                }
            } catch (e) {
                showAlert('error', 'Failed to update row');
            }
        }

        async function deleteRow(row) {
            if (!confirm('Are you sure you want to delete this row?')) return;
            
            try {
                const resp = await fetch('/api/row/' + encodeURIComponent(currentTable) + '/delete/x', {
                    method: 'DELETE',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(row)
                });
                const data = await resp.json();
                if (data.success) {
                    showAlert('success', 'Row deleted successfully');
                    loadTableData();
                    loadTables();
                } else {
                    showAlert('error', data.error);
                }
            } catch (e) {
                showAlert('error', 'Failed to delete row');
            }
        }

        function showImportModal() {
            document.getElementById('importMode').value = 'import';
            document.getElementById('importModalTitle').textContent = 'Import CSV to ' + currentTable;
            document.getElementById('importTableNameGroup').classList.add('hidden');
            document.getElementById('columnCompare').classList.add('hidden');
            document.getElementById('columnEditor').classList.add('hidden');
            document.getElementById('uploadProgress').classList.add('hidden');
            document.getElementById('importBtn').disabled = true;
            document.getElementById('csvFile').value = '';
            document.getElementById('uploadZone').innerHTML = '<p>Drag & drop CSV file here or click to select</p><input type="file" id="csvFile" accept=".csv" style="display:none">';
            rebindFileInput();
            document.getElementById('importModal').classList.add('active');
        }

        function showCreateTableModal() {
            document.getElementById('importMode').value = 'create';
            document.getElementById('importModalTitle').textContent = 'Create Table from CSV';
            document.getElementById('importTableNameGroup').classList.remove('hidden');
            document.getElementById('importTableName').value = '';
            document.getElementById('columnCompare').classList.add('hidden');
            document.getElementById('columnEditor').classList.add('hidden');
            document.getElementById('uploadProgress').classList.add('hidden');
            document.getElementById('importBtn').disabled = true;
            document.getElementById('csvFile').value = '';
            document.getElementById('uploadZone').innerHTML = '<p>Drag & drop CSV file here or click to select</p><input type="file" id="csvFile" accept=".csv" style="display:none">';
            rebindFileInput();
            document.getElementById('importModal').classList.add('active');
        }

        const uploadZone = document.getElementById('uploadZone');
        let csvFile = document.getElementById('csvFile');

        function rebindFileInput() {
            csvFile = document.getElementById('csvFile');
            csvFile.addEventListener('change', handleFileSelect);
        }

        uploadZone.addEventListener('click', () => csvFile.click());
        uploadZone.addEventListener('dragover', e => { e.preventDefault(); uploadZone.classList.add('dragover'); });
        uploadZone.addEventListener('dragleave', () => uploadZone.classList.remove('dragover'));
        uploadZone.addEventListener('drop', e => {
            e.preventDefault();
            uploadZone.classList.remove('dragover');
            if (e.dataTransfer.files.length) {
                csvFile.files = e.dataTransfer.files;
                handleFileSelect();
            }
        });
        csvFile.addEventListener('change', handleFileSelect);

        function handleFileSelect() {
            if (!csvFile.files.length) return;
            const file = csvFile.files[0];
            uploadZone.innerHTML = '<p>Selected: ' + file.name + '</p><input type="file" id="csvFile" accept=".csv" style="display:none">';
            rebindFileInput();

            const mode = document.getElementById('importMode').value;

            const reader = new FileReader();
            reader.onload = e => {
                const firstLine = e.target.result.split(/\r?\n/)[0];
                const headers = firstLine.split(',').map(h => h.trim().replace(/^"|"$/g, ''));

                if (mode === 'create') {
                    if (!document.getElementById('importTableName').value) {
                        document.getElementById('importTableName').value = file.name.replace(/\.csv$/i, '');
                    }
                    const container = document.getElementById('columnInputs');
                    container.innerHTML = headers.map((h, i) =>
                        '<input type="text" class="form-control csv-column-input" data-index="' + i + '" value="' + escapeHtml(h) + '" style="margin-bottom:5px;">'
                    ).join('');
                    document.getElementById('columnEditor').classList.remove('hidden');
                    document.getElementById('columnCompare').classList.add('hidden');
                    document.getElementById('importBtn').disabled = false;
                } else {
                    const tableCols = currentColumns || [];
                    const compareBody = document.getElementById('columnCompareBody');
                    const maxLen = Math.max(headers.length, tableCols.length);
                    let html = '';
                    let matchCount = 0;
                    for (let i = 0; i < maxLen; i++) {
                        const csvCol = headers[i] || '';
                        const tblCol = tableCols[i] || '';
                        const match = csvCol && tblCol && csvCol.toLowerCase() === tblCol.toLowerCase();
                        if (match) matchCount++;
                        html += '<div class="column-compare-row">' +
                            '<span class="csv-col">' + escapeHtml(csvCol || '-') + '</span>' +
                            '<span class="tbl-col">' + escapeHtml(tblCol || '-') + '</span>' +
                            '<span class="status ' + (match ? 'column-match' : 'column-mismatch') + '">' + (match ? '&#10003;' : '&#10007;') + '</span>' +
                            '</div>';
                    }
                    compareBody.innerHTML = html;

                    const summary = document.getElementById('columnCompareSummary');
                    const allMatch = headers.length === tableCols.length && matchCount === headers.length;
                    if (allMatch) {
                        summary.className = 'column-compare-summary ok';
                        summary.textContent = 'Columns match (' + headers.length + ' columns). Ready to import.';
                    } else {
                        summary.className = 'column-compare-summary fail';
                        summary.textContent = 'Columns do not match. CSV has ' + headers.length + ' columns, table has ' + tableCols.length + ' columns.';
                    }
                    document.getElementById('columnCompare').classList.remove('hidden');
                    document.getElementById('columnEditor').classList.add('hidden');
                    document.getElementById('importBtn').disabled = !allMatch;
                }
            };
            reader.readAsText(file);
        }

        async function importCSV() {
            const file = csvFile.files[0];
            if (!file) return;
            
            const mode = document.getElementById('importMode').value;
            const columnNames = [];
            document.querySelectorAll('.csv-column-input').forEach(input => {
                columnNames.push(input.value);
            });
            
            const formData = new FormData();
            formData.append('file', file);
            formData.append('mode', mode);
            if (mode === 'create') {
                formData.append('tableName', document.getElementById('importTableName').value);
                formData.append('columnNames', JSON.stringify(columnNames));
            } else {
                formData.append('tableName', currentTable);
            }
            
            document.getElementById('uploadProgress').classList.remove('hidden');
            document.getElementById('importBtn').disabled = true;
            
            const xhr = new XMLHttpRequest();
            xhr.upload.onprogress = e => {
                if (e.lengthComputable) {
                    const pct = Math.round(e.loaded / e.total * 100);
                    document.getElementById('progressBar').style.width = pct + '%';
                    document.getElementById('progressText').textContent = 'Uploading: ' + pct + '%';
                }
            };
            xhr.onload = () => {
                try {
                    const data = JSON.parse(xhr.responseText);
                    if (data.success) {
                        closeModal('importModal');
                        showAlert('success', 'CSV imported successfully');
                        loadTableData();
                        loadTables();
                    } else {
                        showAlert('error', data.error);
                    }
                } catch (e) {
                    showAlert('error', 'Import failed');
                }
            };
            xhr.onerror = () => showAlert('error', 'Upload failed');
            xhr.open('POST', '/api/import-csv');
            xhr.send(formData);
        }

        function closeModal(id) {
            document.getElementById(id).classList.remove('active');
        }

        function showAlert(type, message) {
            const container = document.getElementById('alertContainer');
            const alert = document.createElement('div');
            alert.className = 'alert alert-' + type;
            alert.textContent = message;
            container.appendChild(alert);
            setTimeout(() => alert.remove(), 5000);
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
</body>
</html>`
