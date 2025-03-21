import React from 'react';
import { Card, Text, Group, ActionIcon, Menu } from '@mantine/core';
import { IconFolder, IconFile, IconDotsVertical, IconDownload, IconTrash, IconShare } from '@tabler/icons-react';

interface FileItem {
  name: string;
  path: string;
  isDirectory: boolean;
  size: number;
  modified: string;
}

interface FileCardProps {
  file: FileItem;
  onClick: () => void;
  onDelete: () => void;
}

const FileCard: React.FC<FileCardProps> = ({ file, onClick, onDelete }) => {
  const formattedDate = new Date(file.modified).toLocaleDateString();
  const formattedSize = file.isDirectory ? '--' : formatFileSize(file.size);
  
  return (
    <Card 
      shadow="sm" 
      padding="lg" 
      radius="md" 
      withBorder
      className="hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <Card.Section className="p-4 flex justify-center items-center" inheritPadding>
        {file.isDirectory ? (
          <IconFolder size={48} color="#4f46e5" />
        ) : (
          <IconFile size={48} color="#6b7280" />
        )}
      </Card.Section>

      <Group justify="space-between" mt="md" mb="xs">
        <Text fw={500} lineClamp={1} className="flex-1">
          {file.name}
        </Text>
        <Menu withinPortal position="bottom-end" shadow="md">
          <Menu.Target>
            <ActionIcon 
              onClick={(e) => { e.stopPropagation(); }}
              variant="subtle"
            >
              <IconDotsVertical size="1rem" />
            </ActionIcon>
          </Menu.Target>

          <Menu.Dropdown>
            {!file.isDirectory && (
              <Menu.Item 
                leftSection={<IconDownload size={14} />}
                onClick={(e) => { e.stopPropagation(); }}
              >
                Download
              </Menu.Item>
            )}
            <Menu.Item 
              leftSection={<IconShare size={14} />}
              onClick={(e) => { e.stopPropagation(); }}
            >
              Share
            </Menu.Item>
            <Menu.Divider />
            <Menu.Item 
              color="red" 
              leftSection={<IconTrash size={14} />}
              onClick={(e) => { e.stopPropagation(); onDelete(); }}
            >
              Delete
            </Menu.Item>
          </Menu.Dropdown>
        </Menu>
      </Group>
      
      <Text size="sm" color="dimmed">
        {formattedDate}
      </Text>
      <Text size="sm" color="dimmed">
        {formattedSize}
      </Text>
    </Card>
  );
};

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default FileCard;