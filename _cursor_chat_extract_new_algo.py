import sqlite3, json, datetime, re
from pathlib import Path

dbs = [
  "c:/Users/pc/AppData/Roaming/Cursor/User/globalStorage/state.vscdb",
  "c:/Users/pc/AppData/Roaming/Cursor/User/globalStorage/state.vscdb.backup",
  "c:/Users/pc/AppData/Roaming/Cursor/User/workspaceStorage/dc445271520997fb21bc6b6bf959f645/state.vscdb",
  "c:/Users/pc/AppData/Roaming/Cursor/User/workspaceStorage/dc445271520997fb21bc6b6bf959f645/state.vscdb.backup",
  "c:/Users/pc/AppData/Roaming/Cursor/User/workspaceStorage/0afd026e73a32465e4117d1502764e06/state.vscdb",
  "c:/Users/pc/AppData/Roaming/Cursor/User/workspaceStorage/0afd026e73a32465e4117d1502764e06/state.vscdb.backup",
]

# 昨天 20:00 本地时间
now = datetime.datetime.now()
cutoff = (now - datetime.timedelta(days=1)).replace(hour=20, minute=0, second=0, microsecond=0)
cutoff_ts = cutoff.timestamp()

# chat-like key 过滤
key_terms = ["chat", "conversation", "composer", "assistant", "anysphere", "cursor"]

# 新算法关键词（你贴的那套）
val_terms = [
  "新算法",
  "DetectMarketStateSingle",
  "DetectMarketStateMultiCycle",
  "MarketStateThresholds",
  "DThreshold",
  "V = (H - L) / delta",
  "(H - L) / delta",
  "权重投票",
  "低波动",
  "高波动",
  "震荡",
  "趋势",
  "delta",
]

out_path = Path("./_cursor_chat_new_algorithm_hits.jsonl")


def decode_value(v):
    if v is None:
        return ""
    if isinstance(v, bytes):
        return v.decode("utf-8", errors="ignore")
    return str(v)


def find_epoch_markers(s: str):
    out = []
    for m in re.findall(r"\b\d{10,13}\b", s):
        try:
            out.append(int(m))
        except Exception:
            pass
    return out


def maybe_after_cutoff(markers):
    for n in markers:
        if n > 10**12 and (n/1000.0) >= cutoff_ts:
            return True
        if 10**9 < n <= 10**12 and float(n) >= cutoff_ts:
            return True
    return False


def main():
    print("cutoff_local=", cutoff.isoformat(" "))

    where_key = " OR ".join(["key LIKE ?" for _ in key_terms])
    params_key = [f"%{t}%" for t in key_terms]

    where_val = " OR ".join(["value LIKE ?" for _ in val_terms])
    params_val = [f"%{t}%" for t in val_terms]

    sql = "SELECT key, value FROM ItemTable WHERE (" + where_key + ") AND (" + where_val + ")"

    total_written = 0
    per_db = {}

    with out_path.open("w", encoding="utf-8") as f:
        for db in dbs:
            p = Path(db)
            if not p.exists():
                continue

            try:
                con = sqlite3.connect(str(p))
                cur = con.cursor()
                tables = [r[0] for r in cur.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").fetchall()]
                if "ItemTable" not in tables:
                    con.close()
                    continue

                rows = cur.execute(sql, params_key + params_val).fetchall()
                per_db[db] = len(rows)

                for k, v in rows:
                    sv = decode_value(v)
                    markers = find_epoch_markers(sv)
                    rec = {
                        "db": db,
                        "key": k,
                        "len": len(sv),
                        "maybe_after_cutoff": maybe_after_cutoff(markers),
                        "time_markers": markers[:20],
                        "snippet": sv[:4000],
                    }
                    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
                    total_written += 1

                con.close()
            except Exception as e:
                per_db[db] = f"error: {e}"

    print("per_db_hits=", per_db)
    print("WROTE", total_written, "records ->", str(out_path.resolve()))


if __name__ == "__main__":
    main()
