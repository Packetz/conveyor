import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  CardActions,
  Button,
  Chip,
  LinearProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  TextField
} from '@mui/material';
import api from '../services/api';

interface Plugin {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  enabled: boolean;
  config: Record<string, any>;
}

const Plugins = () => {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [openConfigDialog, setOpenConfigDialog] = useState(false);
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);

  useEffect(() => {
    fetchPlugins();
  }, []);

  const fetchPlugins = async () => {
    try {
      setLoading(true);

      // In a real implementation, we would fetch this from the API
      // For now, we'll use mock data
      const mockPlugins: Plugin[] = [
        {
          id: '1',
          name: 'Docker Builder',
          version: '1.2.0',
          description: 'Builds and pushes Docker images to a registry',
          author: 'Conveyor Team',
          enabled: true,
          config: {
            registry: 'docker.io',
            username: '${DOCKER_USERNAME}',
            password: '${DOCKER_PASSWORD}'
          }
        },
        {
          id: '2',
          name: 'Security Scanner',
          version: '0.9.5',
          description: 'Scans code for security vulnerabilities',
          author: 'Conveyor Team',
          enabled: true,
          config: {
            severityThreshold: 'HIGH',
            ignorePatterns: ['**/node_modules/**', '**/vendor/**']
          }
        },
        {
          id: '3',
          name: 'Notification Service',
          version: '1.0.0',
          description: 'Sends notifications for pipeline events',
          author: 'Conveyor Team',
          enabled: false,
          config: {
            slackWebhook: 'https://hooks.slack.com/services/xxx/yyy/zzz',
            emailEnabled: true,
            recipients: ['devops@example.com']
          }
        },
        {
          id: '4',
          name: 'AWS Deployer',
          version: '0.8.3',
          description: 'Deploys applications to AWS environments',
          author: 'Community',
          enabled: true,
          config: {
            region: 'us-west-2',
            credentialsPath: '~/.aws/credentials'
          }
        }
      ];

      // We'd normally fetch this with:
      // const response = await api.getPlugins();
      // setPlugins(response);

      setPlugins(mockPlugins);
      setError(null);
    } catch (err) {
      console.error('Error fetching plugins:', err);
      setError('Failed to load plugins. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePlugin = (id: string) => {
    setPlugins(prevPlugins =>
      prevPlugins.map(plugin =>
        plugin.id === id ? { ...plugin, enabled: !plugin.enabled } : plugin
      )
    );
  };

  const handleConfigurePlugin = (plugin: Plugin) => {
    setSelectedPlugin(plugin);
    setOpenConfigDialog(true);
  };

  const handleCloseConfigDialog = () => {
    setOpenConfigDialog(false);
    setSelectedPlugin(null);
  };

  if (loading) {
    return <LinearProgress />;
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3 }}>Plugins</Typography>

      <Grid container spacing={3}>
        {plugins.map((plugin) => (
          <Grid item xs={12} md={6} key={plugin.id}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6">{plugin.name}</Typography>
                  <Chip
                    label={plugin.enabled ? 'Enabled' : 'Disabled'}
                    color={plugin.enabled ? 'success' : 'default'}
                    size="small"
                  />
                </Box>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  {plugin.description}
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, mt: 2 }}>
                  <Chip size="small" label={`v${plugin.version}`} />
                  <Chip size="small" label={`By: ${plugin.author}`} />
                </Box>
              </CardContent>
              <CardActions>
                <Button
                  size="small"
                  variant="outlined"
                  color={plugin.enabled ? 'error' : 'success'}
                  onClick={() => handleTogglePlugin(plugin.id)}
                >
                  {plugin.enabled ? 'Disable' : 'Enable'}
                </Button>
                <Button
                  size="small"
                  variant="outlined"
                  onClick={() => handleConfigurePlugin(plugin)}
                >
                  Configure
                </Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Plugin Configuration Dialog */}
      <Dialog
        open={openConfigDialog}
        onClose={handleCloseConfigDialog}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          {selectedPlugin?.name} Configuration
        </DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            Configure the settings for this plugin.
          </DialogContentText>

          {selectedPlugin && Object.entries(selectedPlugin.config).map(([key, value]) => (
            <TextField
              key={key}
              label={key.charAt(0).toUpperCase() + key.slice(1).replace(/([A-Z])/g, ' $1')}
              value={typeof value === 'string' ? value : JSON.stringify(value)}
              fullWidth
              margin="normal"
              variant="outlined"
            />
          ))}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseConfigDialog}>Cancel</Button>
          <Button onClick={handleCloseConfigDialog} variant="contained" color="primary">
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Plugins; 