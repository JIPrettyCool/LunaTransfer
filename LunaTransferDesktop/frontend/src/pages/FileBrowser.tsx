import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Text, Title, Button, TextInput, Group, Menu, ActionIcon,
  LoadingOverlay, Breadcrumbs, Anchor
} from '@mantine/core';
import { 
  IconUpload, IconSearch, IconFolderPlus, IconDotsVertical,
  IconDownload, IconShare, IconTrash, IconFolder, IconFile, IconArrowLeft
} from '@tabler/icons-react';
import { useAuthStore } from '../store/authStore';
import { ListUserFiles } from '../../wailsjs/go/main/App';

interface FileItem {
  name: string;
  path: string;
  isDirectory: boolean;
  size: number;
  modified: string;
}

const FileBrowser: React.FC = () => {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentPath, setCurrentPath] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const { token } = useAuthStore();

  useEffect(() => {
    loadFiles();
  }, []);

  const loadFiles = async () => {
    if (!token) return;
    
    setLoading(true);
    try {
      const result = await ListUserFiles(token, currentPath);
      setFiles(result.map(item => ({
        name: item.name as string,
        path: item.path as string,
        isDirectory: item.isDirectory as boolean,
        size: item.size as number,
        modified: item.modified as string
      })));
    } catch (error) {
      console.error('Failed to load files:', error);
    } finally {
      setLoading(false);
    }
  };

  const navigateToFolder = (path: string) => {
    setCurrentPath(path);
  };

  const navigateUp = () => {
    const parts = currentPath.split('/');
    parts.pop();
    setCurrentPath(parts.join('/'));
  };

  const filteredFiles = files.filter(file => 
    file.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getBreadcrumbs = () => {
    const parts = currentPath.split('/').filter(p => p);
    const breadcrumbs = [
      { title: 'Home', onClick: () => setCurrentPath('') }
    ];
    
    let path = '';
    parts.forEach(part => {
      path += path ? `/${part}` : part;
      breadcrumbs.push({
        title: part,
        onClick: () => setCurrentPath(path)
      });
    });
    
    return breadcrumbs;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-6 border-b dark:border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <Title order={3}>File Browser</Title>
          
          <Group>
            <Button 
              leftSection={<IconFolderPlus size={16} />}
              variant="outline"
            >
              New Folder
            </Button>
            <Button 
              leftSection={<IconUpload size={16} />}
              className="bg-indigo-600 hover:bg-indigo-700"
            >
              Upload
            </Button>
          </Group>
        </div>
        
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <TextInput
              placeholder="Search files..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />
          </div>
          
          <Breadcrumbs>
            {getBreadcrumbs().map((item, index) => (
              <Anchor 
                key={index} 
                onClick={item.onClick}
                className="cursor-pointer"
              >
                {item.title}
              </Anchor>
            ))}
          </Breadcrumbs>
        </div>
      </div>
      
      <div className="flex-1 overflow-auto p-6 relative">
        <LoadingOverlay 
          visible={loading} 
          loaderProps={{ size: 'md' }}
          zIndex={1000}
        />
        
        {currentPath && (
          <motion.button
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="mb-4 flex items-center text-luna-600 hover:text-luna-700"
            onClick={navigateUp}
          >
            <IconArrowLeft size={16} className="mr-1" />
            <span>Back to parent directory</span>
          </motion.button>
        )}
        
        {filteredFiles.length === 0 && !loading ? (
          <div className="h-full flex items-center justify-center">
            <Text c="dimmed">No files found in this directory</Text>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <AnimatePresence mode="popLayout">
              {filteredFiles.map((file, index) => (
                <motion.div
                  key={file.path}
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.9 }}
                  transition={{ duration: 0.2, delay: index * 0.05 }}
                >
                  <div 
                    className="card hover:bg-gray-50 dark:hover:bg-gray-750 cursor-pointer p-4"
                    onClick={() => file.isDirectory ? navigateToFolder(file.path) : null}
                  >
                    <div className="flex items-center">
                      <div className="mr-3">
                        {file.isDirectory ? (
                          <IconFolder size={32} className="text-yellow-500" />
                        ) : (
                          <IconFile size={32} className="text-blue-500" />
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <Text className="font-medium truncate">{file.name}</Text>
                        <Text size="xs" c="dimmed" className="flex items-center gap-2">
                          {!file.isDirectory && formatFileSize(file.size)} · {new Date(file.modified).toLocaleDateString()}
                        </Text>
                      </div>
                      <Menu position="bottom-end" withArrow>
                        <Menu.Target>
                          <ActionIcon 
                            variant="subtle"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <IconDotsVertical size={16} />
                          </ActionIcon>
                        </Menu.Target>
                        <Menu.Dropdown>
                          {!file.isDirectory && (
                            <Menu.Item leftSection={<IconDownload size={14} />}>
                              Download
                            </Menu.Item>
                          )}
                          <Menu.Item leftSection={<IconShare size={14} />}>
                            Share
                          </Menu.Item>
                          <Menu.Item 
                            leftSection={<IconTrash size={14} />}
                            color="red"
                          >
                            Delete
                          </Menu.Item>
                        </Menu.Dropdown>
                      </Menu>
                    </div>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>
    </div>
  );
};

export default FileBrowser;