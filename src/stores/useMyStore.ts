import create from 'zustand'

export interface MyStore {
  foo: string
}

export const useMyStore = create<MyStore>(() => ({
  foo: 'bar',
}))