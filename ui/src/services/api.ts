import axios from 'axios';

const API_URL = '/api';

// Pipeline types
export interface Pipeline {
  id: string;
  name: string;
  description: string;
  steps: PipelineStep[];
  status: 'idle' | 'running' | 'success' | 'failed';
  createdAt: string;
  updatedAt: string;
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

// API service
const api = {
  // Pipelines
  getPipelines: async (): Promise<Pipeline[]> => {
    const response = await axios.get(`${API_URL}/pipelines`);
    return response.data;
  },

  getPipeline: async (id: string): Promise<Pipeline> => {
    const response = await axios.get(`${API_URL}/pipelines/${id}`);
    return response.data;
  },

  createPipeline: async (pipeline: PipelineCreateDto): Promise<Pipeline> => {
    const response = await axios.post(`${API_URL}/pipelines`, pipeline);
    return response.data;
  },

  updatePipeline: async (id: string, pipeline: Partial<PipelineCreateDto>): Promise<Pipeline> => {
    const response = await axios.put(`${API_URL}/pipelines/${id}`, pipeline);
    return response.data;
  },

  deletePipeline: async (id: string): Promise<void> => {
    await axios.delete(`${API_URL}/pipelines/${id}`);
  },

  runPipeline: async (id: string): Promise<Pipeline> => {
    const response = await axios.post(`${API_URL}/pipelines/${id}/run`);
    return response.data;
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