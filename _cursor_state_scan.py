import sqlite3
from pathlib import Path

dbs = [
  "c:/Users/pc/AppData/Roaming/Cursor/User/globalStorage/state.vscdb",
  "c:/Users/pc/AppData/Roaming/Cursor/User/globalStorage/state.vscdb.backup",
  "c:/Users/pc/AppData/Roaming/Cursor/User/workspaceStorage/dc445271520997fb21bc6b6bf959f645/state.vscdb",
  "c:/Users/pc/AppData/Roaming/Cursor/User/workspaceStorage/dc445271520997fb21bc6b6bf959f645/state.vscdb.backup",
]

for db in dbs:
    p = Path(db)
    if not p.exists():
        continue
    print("\n===", db)
    con = sqlite3.connect(str(p))
    cur = con.cursor()
    tables = [r[0] for r in cur.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").fetchall()]
    print("tables=", tables)
    if 'ItemTable' in tables:
        cnt = cur.execute("SELECT COUNT(1) FROM ItemTable WHERE key LIKE '%chat%' OR key LIKE '%conversation%' OR key LIKE '%composer%' OR key LIKE '%assistant%' OR key LIKE '%anysphere%' OR key LIKE '%cursor%';").fetchone()[0]
        print("ItemTable key hits(chat-like)=", cnt)
    con.close()
