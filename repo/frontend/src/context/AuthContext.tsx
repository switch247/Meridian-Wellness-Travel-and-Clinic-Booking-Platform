import { Alert, Snackbar } from '@mui/material';
import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { api, MeResult } from '../api/client';

type AuthContextValue = {
  token: string | null;
  me: MeResult | null;
  login: (username: string, password: string) => Promise<void>;
  register: (payload: { username: string; password: string; phone: string; address: string }) => Promise<void>;
  logout: () => void;
  refreshMe: () => Promise<void>;
  loading: boolean;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [me, setMe] = useState<MeResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setMe(null);
      return;
    }
    api.me(token)
      .then(setMe)
      .catch((err: Error) => {
        setError(err.message);
        setToken(null);
      });
  }, [token]);

  async function login(username: string, password: string) {
    setLoading(true);
    try {
      const out = await api.login({ username, password });
      setToken(out.token);
      const profile = await api.me(out.token);
      setMe(profile);
    } finally {
      setLoading(false);
    }
  }

  async function register(payload: { username: string; password: string; phone: string; address: string }) {
    setLoading(true);
    try {
      await api.register(payload);
    } finally {
      setLoading(false);
    }
  }

  async function refreshMe() {
    if (!token) return;
    const profile = await api.me(token);
    setMe(profile);
  }

  function logout() {
    setToken(null);
    setMe(null);
  }

  const value = useMemo(
    () => ({ token, me, login, register, logout, refreshMe, loading }),
    [token, me, loading]
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
      <Snackbar open={!!error} autoHideDuration={4000} onClose={() => setError(null)}>
        <Alert severity="error" variant="filled">{error}</Alert>
      </Snackbar>
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
