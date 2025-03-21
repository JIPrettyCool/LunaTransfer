import React, { useState, useEffect } from 'react';
import { Text, Title, Button, TextInput, Group, Modal, LoadingOverlay } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { useAuthStore } from '../store/authStore';
import { 
  ListUserFiles, 
  CreateDirectory, 
  DeleteFile,
  UploadFile 
} from '../../wailsjs/go/main/App';
import FileCard from '../components/FileCard';
import FileUploader from '../components/FileUploader';
import { IconSearch, IconFolderPlus, IconDownload, IconRefresh } from '@tabler/icons-react';

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
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [newFolderName, setNewFolderName] = useState('');
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const { token } = useAuthStore();

  useEffect(() => {
    if (token) {
      loadFiles();
    }
  }, [currentPath, token]);

  const filteredFiles = files.filter(file => 
    file.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const loadFiles = async () => {
    if (!token) {
      console.error("No token available");
      return;
    }
    
    setLoading(true);
    try {
      console.log(`Loading files from path: "${currentPath}"`);
      const result = await ListUserFiles(token, currentPath);
      console.log("Files loaded:", result);
      setFiles(result);
    } catch (error) {
      console.error('Failed to load files:', error);
      
      let errorMessage = 'Failed to load files';
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === 'string') {
        errorMessage = error;
      }
      
      notifications.show({
        title: 'Error',
        message: errorMessage,
        color: 'red',
      });
      
      setFiles([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateFolder = async () => {
    if (!token || !newFolderName.trim()) return;
    
    try {
      await CreateDirectory(token, currentPath, newFolderName);
      notifications.show({
        title: 'Success',
        message: `Folder ${newFolderName} created`,
        color: 'green',
      });
      setNewFolderName('');
      setCreateFolderOpen(false);
      
      setTimeout(() => {
        loadFiles();
      }, 300);
    } catch (error) {
      console.error('Failed to create folder:', error);
      
      let errorMessage = 'Failed to create folder';
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === 'string') {
        errorMessage = error;
      }
      
      notifications.show({
        title: 'Error',
        message: errorMessage,
        color: 'red',
      });
    }
  };

  const handleFileUpload = async (file: File) => {
    if (!token) return;
    
    const reader = new FileReader();
    reader.onload = async (e) => {
      if (!e.target?.result) return;
      
      const fileData = new Uint8Array(e.target.result as ArrayBuffer);
      try {
        await UploadFile(token, currentPath, Array.from(fileData), file.name);
        notifications.show({
          title: 'Success',
          message: `File ${file.name} uploaded`,
          color: 'green',
        });
        
        setTimeout(() => {
          loadFiles();
        }, 300);
      } catch (error) {
        console.error('Failed to upload file:', error);
        
        let errorMessage = 'Failed to upload file';
        if (error instanceof Error) {
          errorMessage = error.message;
        } else if (typeof error === 'string') {
          errorMessage = error;
        }
        
        notifications.show({
          title: 'Error',
          message: errorMessage,
          color: 'red',
        });
      }
    };
    reader.readAsArrayBuffer(file);
  };

  const handleFileClick = (file: FileItem) => {
    if (file.isDirectory) {
      const newPath = file.path;
      console.log(`Navigating to: ${newPath}`);
      setCurrentPath(newPath);
    } else {
      notifications.show({
        title: 'File Selected',
        message: file.name,
        color: 'blue',
      });
    }
  };

  const handleDeleteFile = async (path: string) => {
    if (!token) return;
    
    try {
      await DeleteFile(token, path);
      notifications.show({
        title: 'Success',
        message: 'Item deleted',
        color: 'green',
      });
      
      setTimeout(() => {
        loadFiles();
      }, 300);
    } catch (error) {
      console.error('Failed to delete item:', error);
      
      let errorMessage = 'Failed to delete item';
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === 'string') {
        errorMessage = error;
      }
      
      notifications.show({
        title: 'Error',
        message: errorMessage,
        color: 'red',
      });
    }
  };

  const navigateUp = () => {
    if (!currentPath) return;
    
    const parts = currentPath.split('/');
    parts.pop();
    const parentPath = parts.join('/');
    setCurrentPath(parentPath);
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-6 border-b dark:border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <Title order={3}>Files</Title>
          
          <Group>
            <Button 
              onClick={() => setUploadModalOpen(true)}
              leftSection={<IconDownload size={16} />}
              className="bg-indigo-600 hover:bg-indigo-700"
            >
              Upload
            </Button>
            <Button 
              onClick={() => setCreateFolderOpen(true)}
              leftSection={<IconFolderPlus size={16} />}
              variant="outline"
            >
              New Folder
            </Button>
            <Button onClick={loadFiles} leftSection={<IconRefresh size={16} />}>
              Refresh Files
            </Button>
          </Group>
        </div>
        
        <div className="flex items-center gap-4">
          <TextInput
            placeholder="Search files..."
            leftSection={<IconSearch size={16} />}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.currentTarget.value)}
            className="flex-1"
          />
        </div>
        
        {currentPath && (
          <div className="flex items-center mt-4 text-sm">
            <button onClick={() => setCurrentPath('')} className="hover:underline text-blue-500">
              Root
            </button>
            {currentPath.split('/').filter(Boolean).map((part, index, arr) => {
              const pathToHere = arr.slice(0, index + 1).join('/');
              return (
                <React.Fragment key={index}>
                  <span className="mx-1">/</span>
                  <button 
                    onClick={() => setCurrentPath(pathToHere)} 
                    className="hover:underline text-blue-500"
                  >
                    {part}
                  </button>
                </React.Fragment>
              );
            })}
          </div>
        )}
      </div>
      
      <div className="flex-1 overflow-auto p-6 relative">
        <LoadingOverlay visible={loading} />
        
        {filteredFiles.length === 0 && !loading ? (
          <div className="h-full flex items-center justify-center">
            <Text c="dimmed">No files found</Text>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredFiles.map((file) => (
              <FileCard 
                key={file.path} 
                file={file} 
                onClick={() => handleFileClick(file)}
                onDelete={() => handleDeleteFile(file.path)}
              />
            ))}
          </div>
        )}
      </div>
      
      <Modal 
        opened={createFolderOpen} 
        onClose={() => setCreateFolderOpen(false)}
        title="Create New Folder"
      >
        <TextInput
          label="Folder Name"
          placeholder="Enter folder name"
          value={newFolderName}
          onChange={(e) => setNewFolderName(e.currentTarget.value)}
          className="mb-4"
        />
        <Group justify="right">
          <Button variant="outline" onClick={() => setCreateFolderOpen(false)}>Cancel</Button>
          <Button onClick={handleCreateFolder}>Create</Button>
        </Group>
      </Modal>

      <Modal
        opened={uploadModalOpen}
        onClose={() => setUploadModalOpen(false)}
        title="Upload Files"
        size="lg"
      >
        <FileUploader onUpload={handleFileUpload} />
      </Modal>
    </div>
  );
};

export default FileBrowser;