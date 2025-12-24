SELECT id, pid, title, path, component, permissions FROM hg_admin_menu WHERE path LIKE '%promotion%' OR path LIKE '%invite%' OR path LIKE '%team%' ORDER BY pid, sort;
