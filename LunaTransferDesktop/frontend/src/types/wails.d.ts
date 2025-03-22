interface Window {
  go: {
    main: {
      App: {
        GetDebugSetupInfo: () => Promise<{
          fileExists: boolean;
          filePath: string;
          fileSize: number;
          fileContent: string;
          userCount: number;
          loadError: string;
          setupCompleted: boolean;
        }>;
        CheckSetupStatus: () => Promise<boolean>;
        LoginUser: (username: string, password: string) => Promise<any>;
        PerformSetup: (username: string, password: string, email: string) => Promise<any>;
      };
    };
  };
}