import axios from 'axios';

const API_URL = '/api';

// Pipeline types
export interface Pipeline {
  id: string;
  name: string;
  description: string;
  steps: PipelineStep[];
  stages?: {
    id: string;
    name: string;
    steps: PipelineStep[];
    parallel: boolean;
  }[];
  status: 'idle' | 'running' | 'success' | 'failed';
  createdAt: string;
  updatedAt: string;
}

export interface BackendPipeline {
  id: string;
  name: string;
  description: string;
  stages: {
    id: string;
    name: string;
    steps: BackendPipelineStep[];
    parallel: boolean;
  }[];
  createdAt: string;
  updatedAt: string;
}

export interface BackendPipelineStep {
  id: string;
  name: string;
  type: string;
  command?: string;
  plugin?: string;
  config?: Record<string, any>;
}

export interface PipelineStep {
  id: string;
  name: string;
  type: string;
  config: Record<string, any>;
  status: 'idle' | 'running' | 'success' | 'failed';
}

export interface PipelineCreateDto {
  name: string;
  description: string;
  steps: Omit<PipelineStep, 'id' | 'status'>[];
}

// System stats types
export interface SystemStats {
  cpu: {
    usagePercent: number;
    cores: number;
    modelName?: string;
  };
  memory: {
    total: number;
    used: number;
    free: number;
    usagePercent: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
    usagePercent: number;
    mountPoint: string;
  };
  host: {
    hostname: string;
    platform: string;
    uptime: number;
    bootTime: string;
  };
  timestamp: string;
}

// Helper function to adapt backend pipeline to frontend model
const adaptPipeline = (backendPipeline: BackendPipeline): Pipeline => {
  console.log('Adapting backend pipeline:', backendPipeline);

  // Convert stages into flat steps array
  const steps = backendPipeline.stages.flatMap(stage =>
    stage.steps.map(step => ({
      id: step.id,
      name: step.name,
      type: step.type,
      config: step.config || {},
      status: 'idle' as 'idle' | 'running' | 'success' | 'failed' // Explicitly type the status
    }))
  );

  const result = {
    id: backendPipeline.id,
    name: backendPipeline.name,
    description: backendPipeline.description,
    steps,
    // Preserve the original stages data for UI organization
    stages: backendPipeline.stages.map(stage => ({
      ...stage,
      steps: stage.steps.map(step => ({
        id: step.id,
        name: step.name,
        type: step.type,
        config: step.config || {},
        status: 'idle' as 'idle' | 'running' | 'success' | 'failed'
      }))
    })),
    status: 'idle' as 'idle' | 'running' | 'success' | 'failed', // Explicitly type the status
    createdAt: backendPipeline.createdAt,
    updatedAt: backendPipeline.updatedAt
  };

  console.log('Adapted frontend pipeline:', result);
  return result;
};

// API service
const api = {
  // Pipelines
  getPipelines: async (): Promise<Pipeline[]> => {
    console.log('Fetching pipelines from API');
    const response = await axios.get<BackendPipeline[]>(`${API_URL}/pipelines`);
    console.log('Backend response:', response.data);
    return response.data.map(adaptPipeline);
  },

  getPipeline: async (id: string): Promise<Pipeline> => {
    const response = await axios.get<BackendPipeline>(`${API_URL}/pipelines/${id}`);
    return adaptPipeline(response.data);
  },

  createPipeline: async (pipeline: PipelineCreateDto): Promise<Pipeline> => {
    // Generate a unique ID for the pipeline
    const pipelineId = `pipeline-${Date.now()}-${Math.floor(Math.random() * 1000)}`;

    // Transform the frontend pipeline structure to the backend format
    const backendPipeline = {
      id: pipelineId,
      name: pipeline.name,
      description: pipeline.description,
      stages: [
        {
          id: `stage-${Date.now()}`,
          name: 'Default Stage',
          steps: pipeline.steps.map(step => ({
            id: crypto.randomUUID ? crypto.randomUUID() : `step-${Date.now()}-${Math.floor(Math.random() * 1000)}`,
            name: step.name,
            type: step.type,
            ...(step.type === 'script' && { command: step.config.command }),
            ...(step.type !== 'script' && { plugin: 'security' }),
            config: step.config
          })),
          parallel: false
        }
      ]
    };

    console.log('Sending pipeline create request:', backendPipeline);

    try {
      const response = await axios.post<BackendPipeline>(`${API_URL}/pipelines`, backendPipeline);
      console.log('Create pipeline response:', response.data);
      return adaptPipeline(response.data);
    } catch (error: any) {
      console.error('Error creating pipeline:', error.response?.data || error.message);
      throw error;
    }
  },

  updatePipeline: async (id: string, pipeline: Partial<PipelineCreateDto>): Promise<Pipeline> => {
    try {
      // Fetch the existing pipeline to maintain its structure
      const existingResponse = await axios.get<BackendPipeline>(`${API_URL}/pipelines/${id}`);
      const existingPipeline = existingResponse.data;

      // Create updated backend pipeline
      const backendPipeline = {
        ...existingPipeline,
        id: id, // Make sure ID is explicitly set
        name: pipeline.name || existingPipeline.name,
        description: pipeline.description || existingPipeline.description,
        // If new steps are provided, update the stages
        ...(pipeline.steps && {
          stages: [
            {
              id: `stage-${Date.now()}`,
              name: 'Default Stage',
              steps: pipeline.steps.map(step => ({
                id: crypto.randomUUID ? crypto.randomUUID() : `step-${Date.now()}-${Math.floor(Math.random() * 1000)}`,
                name: step.name,
                type: step.type,
                ...(step.type === 'script' && { command: step.config.command }),
                ...(step.type !== 'script' && { plugin: 'security' }),
                config: step.config
              })),
              parallel: false
            }
          ]
        })
      };

      console.log('Sending pipeline update request:', backendPipeline);

      const response = await axios.put<BackendPipeline>(`${API_URL}/pipelines/${id}`, backendPipeline);
      console.log('Update pipeline response:', response.data);
      return adaptPipeline(response.data);
    } catch (error: any) {
      console.error('Error updating pipeline:', error.response?.data || error.message);
      throw error;
    }
  },

  deletePipeline: async (id: string): Promise<void> => {
    await axios.delete(`${API_URL}/pipelines/${id}`);
  },

  runPipeline: async (id: string): Promise<Pipeline> => {
    try {
      const response = await axios.post<BackendPipeline>(`${API_URL}/pipelines/${id}/execute`);
      return adaptPipeline(response.data);
    } catch (error) {
      console.error('Error executing pipeline:', error);
      // Even if we get an error, return a pipeline object with the execution status
      // This allows the UI to display a "started" message even if the backend
      // doesn't return a full pipeline object
      return {
        id,
        name: 'Pipeline',
        description: 'Execution started',
        steps: [],
        status: 'running',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };
    }
  },

  // System stats
  getSystemStats: async (): Promise<SystemStats> => {
    try {
      console.log("Fetching system stats from API...");
      const response = await axios.get<SystemStats>(`${API_URL}/system/stats`, {
        timeout: 5000, // 5 second timeout
        headers: {
          'Cache-Control': 'no-cache',
          'Pragma': 'no-cache',
          'Expires': '0'
        }
      });

      console.log("Received system stats:", response.data);

      // Validate the data is reasonable
      const data = response.data;
      if (!data || !data.cpu || !data.memory || !data.disk || !data.host) {
        console.error("Invalid system stats data structure:", data);
        throw new Error("Invalid system stats data structure");
      }

      return data;
    } catch (error) {
      console.error('Error fetching system stats:', error);
      // Return mock system stats if the API request fails
      const mockStats: SystemStats = {
        cpu: {
          usagePercent: 25 + (Math.random() * 20),
          cores: 4,
          modelName: 'Virtual CPU (Mock)'
        },
        memory: {
          total: 8 * 1024 * 1024 * 1024, // 8 GB
          used: 3 * 1024 * 1024 * 1024,  // 3 GB
          free: 5 * 1024 * 1024 * 1024,  // 5 GB
          usagePercent: 37.5
        },
        disk: {
          total: 100 * 1024 * 1024 * 1024, // 100 GB
          used: 40 * 1024 * 1024 * 1024,   // 40 GB
          free: 60 * 1024 * 1024 * 1024,   // 60 GB
          usagePercent: 40.0,
          mountPoint: '/'
        },
        host: {
          hostname: 'conveyor-server-mock',
          platform: 'Mock Platform',
          uptime: 86400 * 1000000000, // 1 day in nanoseconds
          bootTime: new Date(Date.now() - 86400000).toISOString() // 1 day ago
        },
        timestamp: new Date().toISOString()
      };
      return mockStats;
    }
  },

  // Plugins
  getPlugins: async (): Promise<any[]> => {
    const response = await axios.get(`${API_URL}/plugins`);
    return response.data;
  },

  // Websocket connection for real-time updates
  getWebSocketURL: (): string => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const host = window.location.host;
    return `${protocol}://${host}/api/ws`;
  }
};

export default api; 