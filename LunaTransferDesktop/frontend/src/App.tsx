"use client";
import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useAuthStore } from './store/authStore';
import { notifications } from '@mantine/notifications';
import { LoginUser } from '../wailsjs/go/main/App';
import LunaLogo from './components/LunaLogo';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';

function App() {
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setLoading(false);
    }, 1500);
    
    return () => clearTimeout(timer);
  }, []);

  if (loading) {
    return (
      <div className="h-screen w-screen flex items-center justify-center bg-gradient-to-br from-indigo-900 to-indigo-700">
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5 }}
          className="text-center"
        >
          <LunaLogo size={120} className="mx-auto animate-float" />
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: '100%' }}
            transition={{ delay: 0.5, duration: 1 }}
            className="h-1 bg-white/30 mt-8 rounded-full overflow-hidden"
          >
            <motion.div
              initial={{ x: '-100%' }}
              animate={{ x: '100%' }}
              transition={{ 
                repeat: Infinity, 
                duration: 1.2,
                ease: "easeInOut" 
              }}
              className="h-full w-1/3 bg-white rounded-full"
            />
          </motion.div>
        </motion.div>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="h-screen w-screen bg-gray-50 dark:bg-gray-900"
    >
      {isAuthenticated ? <Dashboard /> : <Login />}
    </motion.div>
  );
}

export default App;
