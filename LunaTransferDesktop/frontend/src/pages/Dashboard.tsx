import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { ActionIcon, Tooltip } from '@mantine/core';
import { useAuthStore } from '../store/authStore';
import LunaLogo from '../components/LunaLogo';
import FileBrowser from './FileBrowser';

const Dashboard: React.FC = () => {
  const [activeSection, setActiveSection] = useState<'files' | 'share' | 'groups' | 'settings' | 'profile'>('files');
  const [darkMode, setDarkMode] = useState(false);
  const { logout, user } = useAuthStore();
  const isAdmin = (user?.role || 'admin') === 'admin';
  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
    document.documentElement.classList.toggle('dark');
  };

  const renderContent = () => {
    switch (activeSection) {
      case 'files':
        return <FileBrowser />;
      case 'groups':
        return <div className="p-8 text-gray-500">Groups Manager Coming Soon</div>;
      case 'settings':
        return <div className="p-8 text-gray-500">Settings Coming Soon</div>;
      case 'profile':
        return <div className="p-8 text-gray-500">Profile Coming Soon</div>;
      case 'share':
        return <div className="p-8 text-gray-500">Share Manager Coming Soon</div>;
      default:
        return <FileBrowser />;
    }
  };

  return (
    <div className="h-screen flex dark:bg-gray-900">
      {/* Sidebar */}
      <motion.div 
        initial={{ x: -50, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ duration: 0.5 }}
        className="w-16 bg-white dark:bg-gray-800 shadow-lg flex flex-col items-center py-4"
      >
        <div className="mb-8">
          <LunaLogo size={40} />
        </div>
        
        <div className="flex-1 flex flex-col gap-4">
          <Tooltip label="Files" position="right">
            <ActionIcon 
              size="xl" 
              variant={activeSection === 'files' ? 'filled' : 'subtle'}
              color={activeSection === 'files' ? 'indigo' : 'gray'}
              onClick={() => setActiveSection('files')}
              className="transition-all"
            >
              <span className="material-icons">folder</span>
            </ActionIcon>
          </Tooltip>

          <Tooltip label="Share" position="right">
            <ActionIcon 
              size="xl" 
              variant={activeSection === 'share' ? 'filled' : 'subtle'}
              color={activeSection === 'share' ? 'indigo' : 'gray'}
              onClick={() => setActiveSection('share')}
              className="transition-all"
            >
              <span className="material-icons">share</span>
            </ActionIcon>
          </Tooltip>

          {isAdmin && (
          <Tooltip label="Groups" position="right">
            <ActionIcon 
              size="xl" 
              variant={activeSection === 'groups' ? 'filled' : 'subtle'}
              color={activeSection === 'groups' ? 'indigo' : 'gray'}
              onClick={() => setActiveSection('groups')}
              className="transition-all"
            >
              <span className="material-icons">group</span>
            </ActionIcon>
          </Tooltip>
          )}
          {isAdmin && (
          <Tooltip label="Settings" position="right">
            <ActionIcon 
              size="xl" 
              variant={activeSection === 'settings' ? 'filled' : 'subtle'}
              color={activeSection === 'settings' ? 'indigo' : 'gray'}
              onClick={() => setActiveSection('settings')}
              className="transition-all"
            >
              <span className="material-icons">settings</span>
            </ActionIcon>
          </Tooltip>
            )}
          <Tooltip label="Profile" position="right">
            <ActionIcon 
              size="xl" 
              variant={activeSection === 'profile' ? 'filled' : 'subtle'}
              color={activeSection === 'profile' ? 'indigo' : 'gray'}
              onClick={() => setActiveSection('profile')}
              className="transition-all"
            >
              <img src="https://github.com/jiprettycool.png" alt="Profile" className="w-8 h-8 rounded-full" />
            </ActionIcon>
          </Tooltip>
        </div>
        
        <div className="mt-auto flex flex-col gap-4">
          <Tooltip label={darkMode ? "Light Mode" : "Dark Mode"} position="right">
            <ActionIcon 
              size="xl" 
              variant="subtle"
              onClick={toggleDarkMode}
              className="transition-all"
            >
              <span className="material-icons">
                {darkMode ? "light_mode" : "dark_mode"}
              </span>
            </ActionIcon>
          </Tooltip>
          
          <Tooltip label="Logout" position="right">
            <ActionIcon 
              size="xl" 
              variant="subtle"
              color="red"
              onClick={logout}
              className="transition-all"
            >
              <span className="material-icons">logout</span>
            </ActionIcon>
          </Tooltip>
        </div>
      </motion.div>

      {/* Main content */}
      <motion.div 
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.2 }}
        className="flex-1 overflow-hidden"
      >
        {renderContent()}
      </motion.div>
    </div>
  );
};

export default Dashboard;