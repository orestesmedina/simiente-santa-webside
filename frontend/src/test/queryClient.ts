import { QueryClient } from '@tanstack/react-query';

/** QueryClient determinista para pruebas: sin reintentos ni foco. */
export function createQueryClientForTest() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
        staleTime: 0,
        gcTime: 0,
      },
    },
  });
}
