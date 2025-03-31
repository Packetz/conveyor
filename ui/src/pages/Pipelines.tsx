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
  Snackbar
} from '@mui/material';
import {
  Add as AddIcon,
  PlayArrow as PlayArrowIcon,
  Delete as DeleteIcon,
  ExpandMore as ExpandMoreIcon
} from '@mui/icons-material';
import api, { Pipeline, PipelineStep } from '../services/api';

const Pipelines = () => {
  const navigate = useNavigate();
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [newPipeline, setNewPipeline] = useState({
    name: '',
    description: '',
    steps: []
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
      await api.runPipeline(id);
      setSnackbar({
        open: true,
        message: 'Pipeline started',
        severity: 'success'
      });
    } catch (err) {
      console.error('Error running pipeline:', err);
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
                      <Chip
                        label={pipeline.status.toUpperCase()}
                        color={getStatusColor(pipeline.status)}
                        size="small"
                      />
                    </Box>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                      {pipeline.description}
                    </Typography>
                    <Typography variant="caption" display="block" sx={{ mb: 1 }}>
                      Last updated: {new Date(pipeline.updatedAt).toLocaleString()}
                    </Typography>

                    <Accordion>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography>Steps ({pipeline.steps.length})</Typography>
                      </AccordionSummary>
                      <AccordionDetails>
                        {pipeline.steps.map((step: PipelineStep, index: number) => (
                          <Box key={step.id} sx={{ mb: 1, display: 'flex', alignItems: 'center' }}>
                            <Chip
                              size="small"
                              label={`${index + 1}`}
                              sx={{ mr: 1, minWidth: 30 }}
                            />
                            <Typography variant="body2" sx={{ mr: 1 }}>{step.name}</Typography>
                            <Chip
                              size="small"
                              label={step.status.toUpperCase()}
                              color={getStatusColor(step.status)}
                            />
                          </Box>
                        ))}
                      </AccordionDetails>
                    </Accordion>
                  </CardContent>
                  <CardActions>
                    <Button
                      size="small"
                      startIcon={<PlayArrowIcon />}
                      onClick={() => handleRunPipeline(pipeline.id)}
                      disabled={pipeline.status === 'running'}
                    >
                      Run
                    </Button>
                    <Box sx={{ flexGrow: 1 }} />
                    <IconButton
                      size="small"
                      color="error"
                      onClick={() => handleDeletePipeline(pipeline.id)}
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
          />
          {/* TODO: Add UI for configuring pipeline steps */}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button
            onClick={handleCreatePipeline}
            variant="contained"
            disabled={!newPipeline.name}
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