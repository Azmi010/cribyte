import type { FileItem, FolderItem } from "$lib/types";
import { SvelteSet } from "svelte/reactivity";

export interface SelectEntry {
  id: string;
  name: string;
  type: "file" | "folder";
}

/**
 * Reusable multi-select controller (Svelte 5 runes).
 * Menangani: klik tunggal, ctrl/cmd-toggle, shift-range, select-all,
 * dan marquee (rubber-band) selection di area kosong.
 */
export function createSelection(getFolders: () => FolderItem[], getFiles: () => FileItem[]) {
  const selectedIds = new SvelteSet<string>();
  let lastSelectedId = $state<string | null>(null);

  const orderedIds = $derived([...getFolders().map((f) => f.id), ...getFiles().map((f) => f.id)]);

  const singleSelectedId = $derived(selectedIds.size === 1 ? [...selectedIds][0] : null);

  const selectedFilesList = $derived(getFiles().filter((f) => selectedIds.has(f.id)));
  const selectedFoldersList = $derived(getFolders().filter((f) => selectedIds.has(f.id)));

  const selectedEntries = $derived<SelectEntry[]>([
    ...selectedFoldersList.map((f) => ({ id: f.id, name: f.name, type: "folder" as const })),
    ...selectedFilesList.map((f) => ({ id: f.id, name: f.name, type: "file" as const })),
  ]);

  function replace(ids: Iterable<string>) {
    selectedIds.clear();
    for (const id of ids) selectedIds.add(id);
  }

  function clear() {
    selectedIds.clear();
    lastSelectedId = null;
  }

  function handleSelect(id: string, e: MouseEvent) {
    if (e.shiftKey && lastSelectedId) {
      const a = orderedIds.indexOf(lastSelectedId);
      const b = orderedIds.indexOf(id);
      if (a !== -1 && b !== -1) {
        const [lo, hi] = a < b ? [a, b] : [b, a];
        for (let i = lo; i <= hi; i++) selectedIds.add(orderedIds[i]);
      }
      return;
    }
    if (e.ctrlKey || e.metaKey) {
      if (selectedIds.has(id)) selectedIds.delete(id);
      else selectedIds.add(id);
      lastSelectedId = id;
      return;
    }
    if (selectedIds.size === 1 && selectedIds.has(id)) {
      clear();
      return;
    }
    replace([id]);
    lastSelectedId = id;
  }

  function selectAll() {
    if (orderedIds.length === 0) return;
    replace(orderedIds);
    lastSelectedId = orderedIds[orderedIds.length - 1];
  }

  return {
    get selectedIds() {
      return selectedIds;
    },
    get lastSelectedId() {
      return lastSelectedId;
    },
    set lastSelectedId(v: string | null) {
      lastSelectedId = v;
    },
    get orderedIds() {
      return orderedIds;
    },
    get singleSelectedId() {
      return singleSelectedId;
    },
    get selectedFilesList() {
      return selectedFilesList;
    },
    get selectedFoldersList() {
      return selectedFoldersList;
    },
    get selectedEntries() {
      return selectedEntries;
    },
    replace,
    clear,
    handleSelect,
    selectAll,
  };
}

/**
 * Marquee (rubber-band) selection controller.
 * Hasilkan rectangle overlay + hitung item beririsan via [data-select-id].
 */
export function createMarquee(
  getContentEl: () => HTMLElement | null,
  selectedIds: SvelteSet<string>,
  setLastSelectedId: (id: string | null) => void,
  clearSelection: () => void,
) {
  let marquee = $state<{ x: number; y: number; w: number; h: number } | null>(null);
  let start = { x: 0, y: 0 };
  let additive = false;
  let base: string[] = [];

  function onStart(e: MouseEvent) {
    const contentEl = getContentEl();
    if (e.button !== 0 || !contentEl) return;
    const target = e.target as HTMLElement;
    if (target.closest("[data-select-id]") || target.closest("button")) return;

    additive = e.ctrlKey || e.metaKey || e.shiftKey;
    base = additive ? [...selectedIds] : [];
    if (!additive) clearSelection();

    const rect = contentEl.getBoundingClientRect();
    start = {
      x: e.clientX - rect.left + contentEl.scrollLeft,
      y: e.clientY - rect.top + contentEl.scrollTop,
    };
    marquee = { x: start.x, y: start.y, w: 0, h: 0 };

    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onEnd);
  }

  function onMove(e: MouseEvent) {
    const contentEl = getContentEl();
    if (!contentEl || !marquee) return;
    const rect = contentEl.getBoundingClientRect();
    const curX = e.clientX - rect.left + contentEl.scrollLeft;
    const curY = e.clientY - rect.top + contentEl.scrollTop;
    const x = Math.min(start.x, curX);
    const y = Math.min(start.y, curY);
    const w = Math.abs(curX - start.x);
    const h = Math.abs(curY - start.y);
    marquee = { x, y, w, h };

    const boxLeft = x + rect.left - contentEl.scrollLeft;
    const boxTop = y + rect.top - contentEl.scrollTop;
    const boxRight = boxLeft + w;
    const boxBottom = boxTop + h;

    // Set lokal non-reaktif untuk hitung irisan
    // eslint-disable-next-line svelte/prefer-svelte-reactivity
    const next = new Set(base);
    const nodes = contentEl.querySelectorAll<HTMLElement>("[data-select-id]");
    nodes.forEach((node) => {
      const r = node.getBoundingClientRect();
      const intersects =
        r.left < boxRight && r.right > boxLeft && r.top < boxBottom && r.bottom > boxTop;
      const id = node.dataset.selectId;
      if (id && intersects) next.add(id);
    });
    // Sinkronkan isi SvelteSet dengan hasil hitung tanpa membuat instance baru
    for (const id of [...selectedIds]) {
      if (!next.has(id)) selectedIds.delete(id);
    }
    for (const id of next) selectedIds.add(id);
  }

  function onEnd() {
    marquee = null;
    if (selectedIds.size > 0) {
      setLastSelectedId([...selectedIds][selectedIds.size - 1] ?? null);
    }
    window.removeEventListener("mousemove", onMove);
    window.removeEventListener("mouseup", onEnd);
  }

  return {
    get marquee() {
      return marquee;
    },
    onStart,
  };
}
