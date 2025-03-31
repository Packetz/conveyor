import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Typography,
  Button,
  Card,
  CardContent,
  CardActions,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Chip,
  Grid,
  IconButton,
  LinearProgress,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Alert,
  Snackbar,
  CircularProgress
} from '@mui/material';
import {
  Add as AddIcon,
  PlayArrow as PlayArrowIcon,
  Delete as DeleteIcon,
  ExpandMore as ExpandMoreIcon
} from '@mui/icons-material';
import api, { Pipeline, PipelineStep, PipelineCreateDto } from '../services/api';

// Define types for step creation
interface StepForm {
  name: string;
  type: string;
  config: {
    command?: string;
    scanTypes?: string[];
    severityThreshold?: string;
    [key: string]: any;
  };
}

const Pipelines = () => {
  const navigate = useNavigate();
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [newPipeline, setNewPipeline] = useState<PipelineCreateDto>({
    name: '',
    description: '',
    steps: []
  });
  const [newStep, setNewStep] = useState<StepForm>({
    name: '',
    type: 'script',
    config: { command: '' }
  });
  const [snackbar, setSnackbar] = useState({
    open: false,
    message: '',
    severity: 'success' as 'success' | 'error'
  });

  // Fetch pipelines on component mount
  useEffect(() => {
    fetchPipelines();
    setupWebSocket();
  }, []);

  const fetchPipelines = async () => {
    try {
      setLoading(true);
      const data = await api.getPipelines();
      console.log('Fetched pipelines:', data);
      setPipelines(data);
      setError(null);
    } catch (err) {
      console.error('Error fetching pipelines:', err);
      setError('Failed to load pipelines. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const setupWebSocket = () => {
    try {
      const ws = new WebSocket(api.getWebSocketURL());

      ws.onopen = () => {
        console.log('WebSocket connection established');
      };

      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'PIPELINE_UPDATE') {
          // Update the pipeline in the list
          setPipelines(prevPipelines =>
            prevPipelines.map(p =>
              p.id === data.payload.id ? data.payload : p
            )
          );
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onclose = () => {
        console.log('WebSocket connection closed');
        // Attempt to reconnect after a delay
        setTimeout(setupWebSocket, 3000);
      };

      return () => {
        ws.close();
      };
    } catch (err) {
      console.error('Failed to setup WebSocket:', err);
    }
  };

  const handleCreatePipeline = async () => {
    try {
      await api.createPipeline(newPipeline);
      setOpenDialog(false);
      setNewPipeline({ name: '', description: '', steps: [] });
      fetchPipelines();
      setSnackbar({
        open: true,
        message: 'Pipeline created successfully',
        severity: 'success'
      });
    } catch (err) {
      console.error('Error creating pipeline:', err);
      setSnackbar({
        open: true,
        message: 'Failed to create pipeline',
        severity: 'error'
      });
    }
  };

  const handleRunPipeline = async (id: string) => {
    try {
      // Optimistically update the pipeline status
      setPipelines(prevPipelines =>
        prevPipelines.map(p =>
          p.id === id ? { ...p, status: 'running' } : p
        )
      );

      // Call the API to run the pipeline
      await api.runPipeline(id);

      setSnackbar({
        open: true,
        message: 'Pipeline started',
        severity: 'success'
      });
    } catch (err) {
      console.error('Error running pipeline:', err);

      // Revert the optimistic update if there's an error
      fetchPipelines();

      setSnackbar({
        open: true,
        message: 'Failed to run pipeline',
        severity: 'error'
      });
    }
  };

  const handleDeletePipeline = async (id: string) => {
    try {
      await api.deletePipeline(id);
      fetchPipelines();
      setSnackbar({
        open: true,
        message: 'Pipeline deleted',
        severity: 'success'
      });
    } catch (err) {
      console.error('Error deleting pipeline:', err);
      setSnackbar({
        open: true,
        message: 'Failed to delete pipeline',
        severity: 'error'
      });
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'primary';
      case 'success':
        return 'success';
      case 'failed':
        return 'error';
      default:
        return 'default';
    }
  };

  const closeSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Pipelines</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setOpenDialog(true)}
        >
          Create Pipeline
        </Button>
      </Box>

      {/* Debug info */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="subtitle2">Debug Info:</Typography>
        <Typography variant="body2">
          Pipelines loaded: {pipelines.length}
        </Typography>
        <Typography variant="body2">
          Loading state: {loading ? 'Loading...' : 'Done'}
        </Typography>
        {error && (
          <Typography variant="body2" color="error">
            Error: {error}
          </Typography>
        )}

        {/* Data structure warning */}
        {pipelines.length > 0 && !pipelines[0].steps && (
          <Alert severity="warning" sx={{ mt: 1 }}>
            Pipeline data structure mismatch. Expected 'steps' array but got: {JSON.stringify(Object.keys(pipelines[0]))}
          </Alert>
        )}

        {/* Pipeline structure debug */}
        {pipelines.length > 0 && (
          <Box sx={{ mt: 1 }}>
            <Typography variant="body2">
              First pipeline structure:
              ID: {pipelines[0].id},
              Name: {pipelines[0].name},
              Steps: {pipelines[0].steps?.length || 0},
              Stages: {pipelines[0].stages?.length || 0}
            </Typography>
          </Box>
        )}
      </Box>

      {loading ? (
        <LinearProgress />
      ) : error ? (
        <Alert severity="error">{error}</Alert>
      ) : (
        <Grid container spacing={3}>
          {pipelines.length === 0 ? (
            <Grid item xs={12}>
              <Alert severity="info">
                No pipelines found. Create your first pipeline!
              </Alert>
            </Grid>
          ) : (
            pipelines.map((pipeline) => (
              <Grid item xs={12} md={6} lg={4} key={pipeline.id}>
                <Card>
                  <CardContent>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                      <Typography variant="h6">{pipeline.name}</Typography>
                      {pipeline.status && (
                        <Chip
                          label={pipeline.status.toUpperCase()}
                          color={getStatusColor(pipeline.status)}
                          size="small"
                        />
                      )}
                    </Box>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                      {pipeline.description}
                    </Typography>
                    <Typography variant="caption" display="block" sx={{ mb: 1 }}>
                      Last updated: {new Date(pipeline.updatedAt).toLocaleString()}
                    </Typography>

                    <Accordion>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography>
                          {pipeline.stages
                            ? `Stages (${pipeline.stages.length})`
                            : `Steps (${Array.isArray(pipeline.steps) ? pipeline.steps.length : 0})`}
                        </Typography>
                      </AccordionSummary>
                      <AccordionDetails>
                        {pipeline.stages ? (
                          // Display organized by stages
                          pipeline.stages.map((stage, stageIndex) => (
                            <Box key={stage.id} sx={{ mb: 3 }}>
                              <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 'bold' }}>
                                {stageIndex + 1}. {stage.name}
                                {stage.parallel && (
                                  <Chip
                                    size="small"
                                    label="Parallel"
                                    color="info"
                                    variant="outlined"
                                    sx={{ ml: 1, height: 20, fontSize: '0.7rem' }}
                                  />
                                )}
                              </Typography>

                              {stage.steps.map((step, stepIndex) => (
                                <Box
                                  key={step.id}
                                  sx={{
                                    mb: 1,
                                    display: 'flex',
                                    alignItems: 'center',
                                    pl: 2,
                                    borderLeft: '1px dashed #666'
                                  }}
                                >
                                  <Typography variant="body2" sx={{ mr: 1, color: 'text.secondary' }}>
                                    {stageIndex + 1}.{stepIndex + 1}
                                  </Typography>
                                  <Typography variant="body2" sx={{ mr: 1 }}>{step.name}</Typography>
                                  {step.status && (
                                    <Chip
                                      size="small"
                                      label={step.status.toUpperCase()}
                                      color={getStatusColor(step.status)}
                                    />
                                  )}
                                </Box>
                              ))}
                            </Box>
                          ))
                        ) : (
                          // Fallback to flat steps display
                          Array.isArray(pipeline.steps) && pipeline.steps.map((step: PipelineStep, index: number) => (
                            <Box key={step.id || index} sx={{ mb: 1, display: 'flex', alignItems: 'center' }}>
                              <Chip
                                size="small"
                                label={`${index + 1}`}
                                sx={{ mr: 1, minWidth: 30 }}
                              />
                              <Typography variant="body2" sx={{ mr: 1 }}>{step.name}</Typography>
                              {step.status && (
                                <Chip
                                  size="small"
                                  label={step.status.toUpperCase()}
                                  color={getStatusColor(step.status)}
                                />
                              )}
                            </Box>
                          ))
                        )}
                      </AccordionDetails>
                    </Accordion>
                  </CardContent>
                  <CardActions>
                    <Button
                      size="small"
                      startIcon={pipeline.status === 'running' ? null : <PlayArrowIcon />}
                      onClick={() => handleRunPipeline(pipeline.id)}
                      disabled={pipeline.status === 'running'}
                      color="primary"
                      variant={pipeline.status === 'running' ? 'outlined' : 'text'}
                    >
                      {pipeline.status === 'running' ? (
                        <>
                          <Box sx={{ display: 'flex', alignItems: 'center' }}>
                            <Box sx={{ width: 16, height: 16, mr: 1 }}>
                              <CircularProgress size={16} thickness={6} />
                            </Box>
                            Running
                          </Box>
                        </>
                      ) : 'Run'}
                    </Button>
                    <Box sx={{ flexGrow: 1 }} />
                    <IconButton
                      size="small"
                      color="error"
                      onClick={() => handleDeletePipeline(pipeline.id)}
                      disabled={pipeline.status === 'running'}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </CardActions>
                </Card>
              </Grid>
            ))
          )}
        </Grid>
      )}

      {/* Create Pipeline Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New Pipeline</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Pipeline Name"
            fullWidth
            variant="outlined"
            value={newPipeline.name}
            onChange={(e) => setNewPipeline({ ...newPipeline, name: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Description"
            fullWidth
            variant="outlined"
            multiline
            rows={3}
            value={newPipeline.description}
            onChange={(e) => setNewPipeline({ ...newPipeline, description: e.target.value })}
            sx={{ mb: 3 }}
          />

          {/* Pipeline Steps Section */}
          <Typography variant="subtitle1" sx={{ mb: 1 }}>Pipeline Steps</Typography>

          {/* List existing steps */}
          {newPipeline.steps.length > 0 && (
            <Box sx={{ mb: 2 }}>
              {newPipeline.steps.map((step, index) => (
                <Box key={index} sx={{ display: 'flex', alignItems: 'center', mb: 1, p: 1, border: '1px solid #ddd', borderRadius: 1 }}>
                  <Typography variant="body2" sx={{ flexGrow: 1 }}>
                    {index + 1}. {step.name} ({step.type})
                    {step.type === 'script' && step.config.command &&
                      <Typography variant="caption" display="block" color="text.secondary">
                        $ {step.config.command}
                      </Typography>
                    }
                  </Typography>
                  <IconButton
                    size="small"
                    color="error"
                    onClick={() => {
                      const updatedSteps = [...newPipeline.steps];
                      updatedSteps.splice(index, 1);
                      setNewPipeline({ ...newPipeline, steps: updatedSteps });
                    }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Box>
              ))}
            </Box>
          )}

          {/* Add new step form */}
          <Box sx={{ border: '1px solid #ddd', borderRadius: 1, p: 2, mb: 2 }}>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Add New Step</Typography>
            <TextField
              margin="dense"
              label="Step Name"
              fullWidth
              variant="outlined"
              size="small"
              value={newStep.name}
              onChange={(e) => setNewStep({ ...newStep, name: e.target.value })}
              sx={{ mb: 2 }}
            />
            <TextField
              select
              margin="dense"
              label="Step Type"
              fullWidth
              variant="outlined"
              size="small"
              value={newStep.type}
              onChange={(e) => {
                const type = e.target.value;
                // Initialize appropriate config based on type
                let config = {};
                if (type === 'script') {
                  config = { command: '' };
                } else if (type === 'secret-scan' || type === 'vulnerability-scan') {
                  config = { scanTypes: [], severityThreshold: 'MEDIUM' };
                }
                setNewStep({ ...newStep, type, config });
              }}
              sx={{ mb: 2 }}
            >
              <option value="script">Script</option>
              <option value="secret-scan">Secret Scan</option>
              <option value="vulnerability-scan">Vulnerability Scan</option>
            </TextField>

            {/* Dynamic config fields based on step type */}
            {newStep.type === 'script' && (
              <TextField
                margin="dense"
                label="Command"
                fullWidth
                variant="outlined"
                size="small"
                value={newStep.config.command || ''}
                onChange={(e) => setNewStep({
                  ...newStep,
                  config: { ...newStep.config, command: e.target.value }
                })}
              />
            )}

            {(newStep.type === 'secret-scan' || newStep.type === 'vulnerability-scan') && (
              <>
                <TextField
                  select
                  margin="dense"
                  label="Severity Threshold"
                  fullWidth
                  variant="outlined"
                  size="small"
                  value={newStep.config.severityThreshold || 'MEDIUM'}
                  onChange={(e) => setNewStep({
                    ...newStep,
                    config: { ...newStep.config, severityThreshold: e.target.value }
                  })}
                  sx={{ mb: 2 }}
                >
                  <option value="LOW">Low</option>
                  <option value="MEDIUM">Medium</option>
                  <option value="HIGH">High</option>
                  <option value="CRITICAL">Critical</option>
                </TextField>
              </>
            )}

            <Button
              variant="outlined"
              size="small"
              onClick={() => {
                if (newStep.name) {
                  setNewPipeline({
                    ...newPipeline,
                    steps: [...newPipeline.steps, newStep]
                  });
                  // Reset step form
                  setNewStep({
                    name: '',
                    type: 'script',
                    config: { command: '' }
                  });
                }
              }}
              disabled={!newStep.name || (newStep.type === 'script' && !newStep.config.command)}
              sx={{ mt: 1 }}
            >
              Add Step
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button
            onClick={handleCreatePipeline}
            variant="contained"
            disabled={!newPipeline.name || newPipeline.steps.length === 0}
          >
            Create
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={closeSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={closeSnackbar} severity={snackbar.severity}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Pipelines; 