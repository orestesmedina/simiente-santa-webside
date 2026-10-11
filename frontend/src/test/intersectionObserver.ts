/**
 * Doble de `IntersectionObserver` para las pruebas de movimiento (F10).
 * jsdom no implementa `IntersectionObserver`; este stub permite capturar las
 * instancias, ver los elementos observados y disparar intersecciones a mano.
 */
export class IntersectionObserverStub {
  static instances: IntersectionObserverStub[] = [];

  callback: IntersectionObserverCallback;
  targets = new Set<Element>();

  constructor(callback: IntersectionObserverCallback) {
    this.callback = callback;
    IntersectionObserverStub.instances.push(this);
  }

  observe(element: Element): void {
    this.targets.add(element);
  }

  unobserve(element: Element): void {
    this.targets.delete(element);
  }

  disconnect(): void {
    this.targets.clear();
  }

  takeRecords(): IntersectionObserverEntry[] {
    return [];
  }

  /** Dispara el callback con los elementos observados actualmente. */
  trigger(isIntersecting: boolean): void {
    const entries = [...this.targets].map(
      (target) => ({ target, isIntersecting }) as IntersectionObserverEntry,
    );
    this.callback(entries, this as unknown as IntersectionObserver);
  }
}

export function installIntersectionObserver(): typeof IntersectionObserverStub {
  IntersectionObserverStub.instances = [];
  (globalThis as { IntersectionObserver?: unknown }).IntersectionObserver =
    IntersectionObserverStub;
  return IntersectionObserverStub;
}

export function uninstallIntersectionObserver(): void {
  Reflect.deleteProperty(globalThis, 'IntersectionObserver');
  IntersectionObserverStub.instances = [];
}
