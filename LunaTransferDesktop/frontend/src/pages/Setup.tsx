import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { TextInput, PasswordInput, Button, Text, Title } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { IconUser, IconLock, IconMail, IconDeviceFloppy } from '@tabler/icons-react';
import { PerformSetup } from '../../wailsjs/go/main/App';
import { useAuthStore } from '../store/authStore';
import LunaLogo from '../components/LunaLogo';

const Setup: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [setupComplete, setSetupComplete] = useState(false);
  const { setSetupStatus } = useAuthStore();

  const handleSetup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password || !email) {
      notifications.show({
        title: 'Error',
        message: 'Please fill in all fields',
        color: 'red',
      });
      return;
    }

    setLoading(true);
    try {
      const result = await PerformSetup(username, password, email);
      console.log("Setup result:", result);
      
      if (result && result.success) {
        notifications.show({
          title: 'Success',
          message: 'Setup completed successfully! Please log in with your new credentials.',
          color: 'green',
        });
      
        setSetupStatus(true);
        setSetupComplete(true);
        
        setTimeout(() => {
          window.location.reload();
        }, 2000);
      }
    } catch (error: any) {
      console.error("Setup error:", error);
      notifications.show({
        title: 'Setup Failed',
        message: error.message || 'Could not complete system setup',
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
                Initial Setup
              </Title>
              <Text size="sm" c="dimmed">
                Create Admin Account
              </Text>
            </motion.div>
          </div>
          
          {setupComplete ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-4"
            >
              <Text size="lg" fw={500} className="mb-2">
                Setup completed successfully!
              </Text>
              <Text size="sm" c="dimmed" className="mb-4">
                Redirecting to login page...
              </Text>
              <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: '100%' }}
                  transition={{ duration: 2 }}
                  className="h-full bg-indigo-500"
                />
              </div>
            </motion.div>
          ) : (
            <form onSubmit={handleSetup}>
              <motion.div
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4, duration: 0.5 }}
              >
                <TextInput
                  label="Username"
                  placeholder="Enter admin username"
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
                  placeholder="Enter secure password"
                  leftSection={<IconLock size={16} />}
                  value={password}
                  onChange={(e) => setPassword(e.currentTarget.value)}
                  required
                  className="mb-4"
                />
              </motion.div>
              
              <motion.div
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.6, duration: 0.5 }}
              >
                <TextInput
                  label="Email"
                  placeholder="Enter admin email"
                  leftSection={<IconMail size={16} />}
                  value={email}
                  onChange={(e) => setEmail(e.currentTarget.value)}
                  required
                  className="mb-6"
                />
              </motion.div>
              
              <motion.div
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.7, duration: 0.5 }}
                className="flex justify-center"
              >
                <Button 
                  type="submit"
                  loading={loading}
                  leftSection={<IconDeviceFloppy size={16} />}
                  className="bg-indigo-600 hover:bg-indigo-700 w-full"
                >
                  Complete Setup
                </Button>
              </motion.div>
            </form>
          )}
          
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.8, duration: 0.5 }}
            className="mt-4 text-center"
          >
            <Text size="sm" c="dimmed">
              This will create an administrator account for LunaTransfer
            </Text>
          </motion.div>
        </div>
      </motion.div>
    </div>
  );
};

export default Setup;