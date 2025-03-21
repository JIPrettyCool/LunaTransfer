import React, { useRef, useState } from 'react';
import { Group, Text, Button, rem } from '@mantine/core';
import { Dropzone, FileWithPath } from '@mantine/dropzone';
import { IconUpload, IconX, IconFile } from '@tabler/icons-react';

interface FileUploaderProps {
  onUpload: (file: File) => void;
}

const FileUploader: React.FC<FileUploaderProps> = ({ onUpload }) => {
  const [uploading, setUploading] = useState(false);
  const openRef = useRef<() => void>(null);

  const handleDrop = async (files: FileWithPath[]) => {
    if (files.length === 0) return;
    
    setUploading(true);
    
    try {
      for (const file of files) {
        await onUpload(file);
      }
    } finally {
      setUploading(false);
    }
  };

  return (
    <>
      <Dropzone
        onDrop={handleDrop}
        openRef={openRef}
        loading={uploading}
        className="border-2 border-dashed p-8 rounded-md mb-4"
        maxSize={100 * 1024 * 1024} // 100MB max size
      >
        <Group justify="center" gap="xl" style={{ minHeight: rem(140), pointerEvents: 'none' }}>
          <Dropzone.Accept>
            <IconUpload
              size="3.2rem"
              stroke={1.5}
              color="#4f46e5"
            />
          </Dropzone.Accept>
          <Dropzone.Reject>
            <IconX
              size="3.2rem"
              stroke={1.5}
              color="#ef4444"
            />
          </Dropzone.Reject>
          <Dropzone.Idle>
            <IconFile size="3.2rem" stroke={1.5} />
          </Dropzone.Idle>

          <div>
            <Text size="xl" inline>
              Drag files here or click to select
            </Text>
            <Text size="sm" c="dimmed" inline mt={7}>
              Attach up to 5 files, each file should not exceed 100MB
            </Text>
          </div>
        </Group>
      </Dropzone>

      <Group justify="right">
        <Button variant="outline" onClick={() => openRef.current?.()}>
          Select Files
        </Button>
      </Group>
    </>
  );
};

export default FileUploader;