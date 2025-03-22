import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface UserInfo {
  username: string;
  role: string;
}

interface AuthState {
  token: string | null;
  userInfo: UserInfo | null;
  isAuthenticated: boolean;
  setupCompleted: boolean;
  login: (token: string, userInfo: UserInfo) => void;
  logout: () => void;
  setSetupStatus: (completed: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userInfo: null,
      isAuthenticated: false,
      setupCompleted: false,
      login: (token, userInfo) => set({ token, userInfo, isAuthenticated: true }),
      logout: () => set({ token: null, userInfo: null, isAuthenticated: false }),
      setSetupStatus: (completed) => set({ setupCompleted: completed }),
    }),
    {
      name: 'auth-storage',
    }
  )
);