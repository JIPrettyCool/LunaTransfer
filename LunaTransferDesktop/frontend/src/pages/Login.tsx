import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { TextInput, PasswordInput, Button, Text, Title } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { IconUser, IconLock, IconLogin } from '@tabler/icons-react';
import { LoginUser } from '../../wailsjs/go/main/App';
import { useAuthStore } from '../store/authStore';
import LunaLogo from '../components/LunaLogo';

const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuthStore();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password) {
      notifications.show({
        title: 'Error',
        message: 'Please enter both username and password',
        color: 'red',
      });
      return;
    }

    setLoading(true);
    try {
      const result = await LoginUser(username, password);
      login(result.token as string, result.username as string);
      notifications.show({
        title: 'Success',
        message: 'Logged in successfully!',
        color: 'green',
      });
    } catch (error: any) {
      notifications.show({
        title: 'Login Failed',
        message: error.message || 'Invalid username or password',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-indigo-900 via-indigo-800 to-indigo-700">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl overflow-hidden max-w-md w-full"
      >
        <div className="p-8">
          <div className="flex items-center justify-center mb-8">
            <LunaLogo size={60} />
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.3, duration: 0.5 }}
              className="ml-4"
            >
              <Title order={1} className="text-indigo-600 dark:text-indigo-400">
                LunaTransfer
              </Title>
              <Text size="sm" c="dimmed">
                Secure File Management
              </Text>
            </motion.div>
          </div>
          
          <form onSubmit={handleLogin}>
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 0.5 }}
            >
              <TextInput
                label="Username"
                placeholder="Enter your username"
                leftSection={<IconUser size={16} />}
                value={username}
                onChange={(e) => setUsername(e.currentTarget.value)}
                required
                className="mb-4"
              />
            </motion.div>
            
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.5, duration: 0.5 }}
            >
              <PasswordInput
                label="Password"
                placeholder="Enter your password"
                leftSection={<IconLock size={16} />}
                value={password}
                onChange={(e) => setPassword(e.currentTarget.value)}
                required
                className="mb-6"
              />
            </motion.div>
            
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.6, duration: 0.5 }}
              className="flex justify-center"
            >
              <Button 
                type="submit"
                loading={loading}
                leftSection={<IconLogin size={16} />}
                className="bg-indigo-600 hover:bg-indigo-700 w-full"
              >
                Login
              </Button>
            </motion.div>
          </form>
        </div>
        
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.7, duration: 0.5 }}
          className="bg-gray-100 dark:bg-gray-700 p-4 text-center"
        >
          <Text size="sm" c="dimmed">
            © {new Date().getFullYear()} LunaTransfer • Secure File Management
          </Text>
        </motion.div>
      </motion.div>
    </div>
  );
};

export default Login;