export interface TabItem {
  id: string;
  label: string;
}

export interface TabsProps {
  tabs: TabItem[];
  active: string;
  onChange: (id: string) => void;
  /** Nombre accesible del conjunto de pestañas. */
  label?: string;
}

/**
 * Pestañas con texto visible y estado en `aria-current="page"` (ux.md §6), sin
 * depender solo del color. Las usa la Auditoría para sus dos historiales
 * (ux.md §3.10).
 */
export function Tabs({ tabs, active, onChange, label = 'Pestañas' }: TabsProps) {
  return (
    <nav aria-label={label} className="flex flex-wrap gap-2 border-b border-slate-200">
      {tabs.map((tab) => {
        const selected = tab.id === active;
        const classes = [
          'min-h-11 rounded-t px-4 py-2 font-medium',
          'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
          selected
            ? 'border-b-2 border-slate-900 font-semibold text-slate-900'
            : 'text-slate-600 hover:text-slate-900',
        ].join(' ');

        return (
          <button
            key={tab.id}
            type="button"
            aria-current={selected ? 'page' : undefined}
            onClick={() => onChange(tab.id)}
            className={classes}
          >
            {tab.label}
          </button>
        );
      })}
    </nav>
  );
}
