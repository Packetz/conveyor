import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Grid,
  Paper,
  Card,
  CardContent,
  CardHeader,
  List,
  ListItem,
  ListItemText,
  Divider,
  LinearProgress,
  Alert
} from '@mui/material';
import {
  Assessment as AssessmentIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  Speed as SpeedIcon
} from '@mui/icons-material';
import api, { Pipeline } from '../services/api';

interface SystemMetrics {
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  uptime: number;
  activeJobs: number;
  queuedJobs: number;
  totalPipelines: number;
  pipelinesRunToday: number;
}

interface RecentActivity {
  timestamp: string;
  message: string;
  type: 'success' | 'error' | 'info' | 'warning';
}

const Dashboard = () => {
  const [metrics, setMetrics] = useState<SystemMetrics>({
    cpuUsage: 0,
    memoryUsage: 0,
    diskUsage: 0,
    uptime: 0,
    activeJobs: 0,
    queuedJobs: 0,
    totalPipelines: 0,
    pipelinesRunToday: 0
  });
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([]);
  const [recentPipelines, setRecentPipelines] = useState<Pipeline[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);

        // In a real implementation, you would fetch actual metrics from the API
        // For now, we'll generate mock data
        const mockMetrics: SystemMetrics = {
          cpuUsage: Math.random() * 100,
          memoryUsage: Math.random() * 100,
          diskUsage: Math.random() * 100,
          uptime: Math.floor(Math.random() * 100000),
          activeJobs: Math.floor(Math.random() * 10),
          queuedJobs: Math.floor(Math.random() * 20),
          totalPipelines: Math.floor(Math.random() * 100),
          pipelinesRunToday: Math.floor(Math.random() * 50)
        };

        const mockActivity: RecentActivity[] = [
          {
            timestamp: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
            message: 'Pipeline "Frontend Build" completed successfully',
            type: 'success'
          },
          {
            timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
            message: 'Pipeline "Backend Tests" failed during step "Integration Tests"',
            type: 'error'
          },
          {
            timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
            message: 'New plugin "Docker Builder" installed',
            type: 'info'
          },
          {
            timestamp: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
            message: 'System updated to version 1.2.0',
            type: 'info'
          },
          {
            timestamp: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
            message: 'Low disk space warning (85% used)',
            type: 'warning'
          }
        ];

        // Fetch actual pipelines from the API
        const pipelines = await api.getPipelines();

        setMetrics(mockMetrics);
        setRecentActivity(mockActivity);
        setRecentPipelines(pipelines.slice(0, 5)); // Show only the 5 most recent
        setError(null);
      } catch (err) {
        console.error('Error fetching dashboard data:', err);
        setError('Failed to load dashboard data. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    // In a real app, you might want to set up an interval to refresh this data
    const intervalId = setInterval(fetchData, 30000);

    return () => clearInterval(intervalId);
  }, []);

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    return `${days}d ${hours}h ${minutes}m`;
  };

  const MetricCard = ({ title, value, icon, color }: { title: string, value: string | number, icon: React.ReactNode, color: string }) => (
    <Paper elevation={2} sx={{ height: '100%' }}>
      <Box sx={{ p: 2, display: 'flex', alignItems: 'center' }}>
        <Box sx={{
          backgroundColor: `${color}.light`,
          color: `${color}.dark`,
          p: 1.5,
          borderRadius: 2,
          mr: 2
        }}>
          {icon}
        </Box>
        <Box>
          <Typography variant="h4" component="div">
            {value}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {title}
          </Typography>
        </Box>
      </Box>
    </Paper>
  );

  if (loading) {
    return <LinearProgress />;
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3 }}>Dashboard</Typography>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="CPU Usage"
            value={`${metrics.cpuUsage.toFixed(1)}%`}
            icon={<SpeedIcon />}
            color="primary"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Memory Usage"
            value={`${metrics.memoryUsage.toFixed(1)}%`}
            icon={<MemoryIcon />}
            color="secondary"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Disk Usage"
            value={`${metrics.diskUsage.toFixed(1)}%`}
            icon={<StorageIcon />}
            color="warning"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Uptime"
            value={formatUptime(metrics.uptime)}
            icon={<AssessmentIcon />}
            color="success"
          />
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardHeader title="System Stats" />
            <CardContent>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center', bgcolor: 'primary.light', color: 'primary.contrastText' }}>
                    <Typography variant="h5">{metrics.activeJobs}</Typography>
                    <Typography variant="body2">Active Jobs</Typography>
                  </Paper>
                </Grid>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center', bgcolor: 'secondary.light', color: 'secondary.contrastText' }}>
                    <Typography variant="h5">{metrics.queuedJobs}</Typography>
                    <Typography variant="body2">Queued Jobs</Typography>
                  </Paper>
                </Grid>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center', bgcolor: 'info.light', color: 'info.contrastText' }}>
                    <Typography variant="h5">{metrics.totalPipelines}</Typography>
                    <Typography variant="body2">Total Pipelines</Typography>
                  </Paper>
                </Grid>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center', bgcolor: 'success.light', color: 'success.contrastText' }}>
                    <Typography variant="h5">{metrics.pipelinesRunToday}</Typography>
                    <Typography variant="body2">Runs Today</Typography>
                  </Paper>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardHeader title="Recent Activity" />
            <CardContent sx={{ maxHeight: 300, overflow: 'auto' }}>
              <List>
                {recentActivity.map((activity, index) => (
                  <Box key={index}>
                    <ListItem>
                      <ListItemText
                        primary={activity.message}
                        secondary={new Date(activity.timestamp).toLocaleString()}
                        primaryTypographyProps={{
                          color: activity.type === 'error' ? 'error.main' :
                            activity.type === 'warning' ? 'warning.main' :
                              activity.type === 'success' ? 'success.main' : 'inherit'
                        }}
                      />
                    </ListItem>
                    {index < recentActivity.length - 1 && <Divider />}
                  </Box>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card>
            <CardHeader title="Recent Pipelines" />
            <CardContent>
              {recentPipelines.length === 0 ? (
                <Alert severity="info">No pipelines have been created yet.</Alert>
              ) : (
                <List>
                  {recentPipelines.map((pipeline, index) => (
                    <Box key={pipeline.id}>
                      <ListItem>
                        <ListItemText
                          primary={pipeline.name}
                          secondary={`Status: ${pipeline.status} | Last updated: ${new Date(pipeline.updatedAt).toLocaleString()}`}
                          primaryTypographyProps={{
                            fontWeight: 'medium'
                          }}
                          secondaryTypographyProps={{
                            color: pipeline.status === 'failed' ? 'error.main' :
                              pipeline.status === 'success' ? 'success.main' : 'inherit'
                          }}
                        />
                      </ListItem>
                      {index < recentPipelines.length - 1 && <Divider />}
                    </Box>
                  ))}
                </List>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard; 