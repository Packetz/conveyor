import React, { useState, useEffect } from 'react';
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
  Alert,
  Tooltip,
  CircularProgress,
  Chip,
  IconButton,
  Collapse
} from '@mui/material';
import {
  Assessment as AssessmentIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  Speed as SpeedIcon,
  Computer as ComputerIcon,
  PlayArrow as PlayArrowIcon,
  Timeline as TimelineIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  Warning as WarningIcon
} from '@mui/icons-material';
import api, { Pipeline, SystemStats } from '../services/api';

interface SystemMetrics {
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  uptime: number;
  hostname: string;
  platform: string;
  cpuModel: string;
  memoryTotal: number;
  diskTotal: number;
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
    hostname: '',
    platform: '',
    cpuModel: '',
    memoryTotal: 0,
    diskTotal: 0,
    activeJobs: 0,
    queuedJobs: 0,
    totalPipelines: 0,
    pipelinesRunToday: 0
  });
  const [systemStats, setSystemStats] = useState<SystemStats | null>(null);
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([]);
  const [recentPipelines, setRecentPipelines] = useState<Pipeline[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Collapse state for each section
  const [sectionsCollapsed, setSectionsCollapsed] = useState({
    systemResources: false,
    systemInfo: false,
    pipelineStats: false
  });

  // Add state for tracking if data is real or mock
  const [usingMockData, setUsingMockData] = useState(false);

  // Toggle section collapse
  const toggleSection = (section: keyof typeof sectionsCollapsed) => {
    setSectionsCollapsed(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null);

      try {
        // Fetch system stats
        let systemStatsData: SystemStats;
        try {
          console.log("Fetching system stats from backend...");
          systemStatsData = await api.getSystemStats();
          console.log("Received system stats:", systemStatsData);
          setSystemStats(systemStatsData);

          // Check if data might be mock data by looking for mock indicators
          const isMockData =
            systemStatsData.cpu.modelName?.includes('Mock') ||
            systemStatsData.host.hostname?.includes('mock') ||
            systemStatsData.host.platform?.includes('Mock');

          console.log("Using mock data:", isMockData);
          setUsingMockData(isMockData);
        } catch (err) {
          console.error('Error fetching system stats:', err);
          setError('Failed to fetch system metrics');
          setUsingMockData(true);
          return;
        }

        // Update metrics
        setMetrics({
          cpuUsage: systemStatsData.cpu.usagePercent,
          memoryUsage: systemStatsData.memory.usagePercent,
          diskUsage: systemStatsData.disk.usagePercent,
          uptime: systemStatsData.host.uptime / 1000000000, // Convert nanoseconds to seconds
          hostname: systemStatsData.host.hostname,
          platform: systemStatsData.host.platform,
          cpuModel: systemStatsData.cpu.modelName || 'Unknown CPU',
          memoryTotal: systemStatsData.memory.total,
          diskTotal: systemStatsData.disk.total,
          activeJobs: 2,
          queuedJobs: 1,
          totalPipelines: 5,
          pipelinesRunToday: 3
        });

        // Fetch pipelines data
        try {
          const pipelines = await api.getPipelines();

          setMetrics(prevMetrics => ({
            ...prevMetrics,
            totalPipelines: pipelines.length,
            // Count pipelines that were run today
            pipelinesRunToday: pipelines.filter(p => {
              const updatedDate = new Date(p.updatedAt);
              const today = new Date();
              return updatedDate.getDate() === today.getDate() &&
                updatedDate.getMonth() === today.getMonth() &&
                updatedDate.getFullYear() === today.getFullYear();
            }).length
          }));

          setRecentPipelines(pipelines.slice(0, 5)); // Show only the 5 most recent
        } catch (err) {
          console.error('Error fetching pipelines:', err);
          // Use empty array if fetch fails
          setRecentPipelines([]);
        }

        // Mock activity data - in a real app, this would come from an API
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
        setRecentActivity(mockActivity);

      } catch (err) {
        console.error('Error fetching data:', err);
        setError('Failed to fetch data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    // Refresh data every 30 seconds instead of 10 to reduce load
    const intervalId = setInterval(fetchData, 30000);

    return () => clearInterval(intervalId);
  }, []);

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    return `${days}d ${hours}h ${minutes}m`;
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  const MetricCard = ({ title, value, icon, color, subtitle }: {
    title: string,
    value: string | number,
    icon: React.ReactNode,
    color: string,
    subtitle?: string
  }) => (
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
          <Typography variant="h5" component="div">
            {value}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {title}
          </Typography>
          {subtitle && (
            <Typography variant="caption" color="text.secondary" display="block">
              {subtitle}
            </Typography>
          )}
        </Box>
      </Box>
    </Paper>
  );

  const ResourceUsageCard = ({ title, value, total, icon, color }: {
    title: string,
    value: number,
    total: number,
    icon: React.ReactNode,
    color: string
  }) => (
    <Paper elevation={2} sx={{ height: '100%' }}>
      <Box sx={{ p: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          <Box sx={{
            backgroundColor: `${color}.light`,
            color: `${color}.dark`,
            p: 1,
            borderRadius: 2,
            mr: 2
          }}>
            {icon}
          </Box>
          <Typography variant="h6">{title}</Typography>
        </Box>

        <Box sx={{ position: 'relative', display: 'inline-flex', width: '100%', justifyContent: 'center', my: 1 }}>
          <CircularProgress
            variant="determinate"
            value={value > 100 ? 100 : value}
            size={80}
            thickness={5}
            color={value > 90 ? "error" : value > 70 ? "warning" : "success"}
          />
          <Box
            sx={{
              top: 0,
              left: 0,
              bottom: 0,
              right: 0,
              position: 'absolute',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Tooltip title={`${formatBytes(total * value / 100)} used of ${formatBytes(total)}`}>
              <Typography variant="caption" component="div" color="text.secondary" sx={{ fontWeight: 'bold' }}>
                {`${Math.round(value)}%`}
              </Typography>
            </Tooltip>
          </Box>
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

      {/* Mock data warning */}
      {usingMockData && (
        <Alert
          severity="info"
          icon={<WarningIcon />}
          sx={{ mb: 3 }}
          action={
            <IconButton
              color="inherit"
              size="small"
              onClick={() => setUsingMockData(false)}
              aria-label="dismiss mock data message"
            >
              <ExpandLessIcon />
            </IconButton>
          }
        >
          Showing mock system metrics - real-time system stats could not be obtained from the backend.
        </Alert>
      )}

      {/* System Stats Section */}
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">System Resources</Typography>
        <IconButton
          onClick={() => toggleSection('systemResources')}
          aria-expanded={!sectionsCollapsed.systemResources}
          aria-label="toggle system resources"
          size="small"
          sx={{ ml: 1 }}
        >
          {sectionsCollapsed.systemResources ? <ExpandMoreIcon /> : <ExpandLessIcon />}
        </IconButton>
      </Box>
      <Collapse in={!sectionsCollapsed.systemResources} timeout="auto" unmountOnExit>
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={4}>
            <ResourceUsageCard
              title="CPU Usage"
              value={metrics.cpuUsage}
              total={100 * metrics.cpuUsage} // Just for display
              icon={<SpeedIcon />}
              color="primary"
            />
          </Grid>
          <Grid item xs={12} sm={4}>
            <ResourceUsageCard
              title="Memory Usage"
              value={metrics.memoryUsage}
              total={metrics.memoryTotal}
              icon={<MemoryIcon />}
              color="secondary"
            />
          </Grid>
          <Grid item xs={12} sm={4}>
            <ResourceUsageCard
              title="Disk Usage"
              value={metrics.diskUsage}
              total={metrics.diskTotal}
              icon={<StorageIcon />}
              color="warning"
            />
          </Grid>
        </Grid>
      </Collapse>

      {/* System Info */}
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">System Information</Typography>
        <IconButton
          onClick={() => toggleSection('systemInfo')}
          aria-expanded={!sectionsCollapsed.systemInfo}
          aria-label="toggle system information"
          size="small"
          sx={{ ml: 1 }}
        >
          {sectionsCollapsed.systemInfo ? <ExpandMoreIcon /> : <ExpandLessIcon />}
        </IconButton>
      </Box>
      <Collapse in={!sectionsCollapsed.systemInfo} timeout="auto" unmountOnExit>
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Uptime"
              value={formatUptime(metrics.uptime)}
              icon={<TimelineIcon />}
              color="info"
            />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Hostname"
              value={metrics.hostname}
              icon={<ComputerIcon />}
              color="success"
              subtitle={metrics.platform}
            />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="CPU"
              value={`${systemStats?.cpu.cores || 0} Cores`}
              icon={<SpeedIcon />}
              color="primary"
              subtitle={metrics.cpuModel}
            />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Total Memory"
              value={formatBytes(metrics.memoryTotal)}
              icon={<MemoryIcon />}
              color="secondary"
            />
          </Grid>
        </Grid>
      </Collapse>

      {/* Pipeline Stats */}
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">Pipeline Statistics</Typography>
        <IconButton
          onClick={() => toggleSection('pipelineStats')}
          aria-expanded={!sectionsCollapsed.pipelineStats}
          aria-label="toggle pipeline statistics"
          size="small"
          sx={{ ml: 1 }}
        >
          {sectionsCollapsed.pipelineStats ? <ExpandMoreIcon /> : <ExpandLessIcon />}
        </IconButton>
      </Box>
      <Collapse in={!sectionsCollapsed.pipelineStats} timeout="auto" unmountOnExit>
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Total Pipelines"
              value={metrics.totalPipelines}
              icon={<AssessmentIcon />}
              color="primary"
            />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Pipelines Run Today"
              value={metrics.pipelinesRunToday}
              icon={<PlayArrowIcon />}
              color="success"
            />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Active Jobs"
              value={metrics.activeJobs}
              icon={<PlayArrowIcon />}
              color="warning"
            />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <MetricCard
              title="Queued Jobs"
              value={metrics.queuedJobs}
              icon={<TimelineIcon />}
              color="info"
            />
          </Grid>
        </Grid>
      </Collapse>

      <Grid container spacing={3}>
        {/* Recent Pipelines */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardHeader title="Recent Pipelines" />
            <CardContent>
              {recentPipelines.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                  No pipelines found
                </Typography>
              ) : (
                <List>
                  {recentPipelines.map((pipeline, index) => (
                    <React.Fragment key={pipeline.id}>
                      <ListItem>
                        <ListItemText
                          primary={pipeline.name}
                          secondary={`Last updated: ${new Date(pipeline.updatedAt).toLocaleString()}`}
                        />
                        <Chip
                          label={pipeline.status.toUpperCase()}
                          color={
                            pipeline.status === 'success' ? 'success' :
                              pipeline.status === 'failed' ? 'error' :
                                pipeline.status === 'running' ? 'primary' : 'default'
                          }
                          size="small"
                        />
                      </ListItem>
                      {index < recentPipelines.length - 1 && <Divider />}
                    </React.Fragment>
                  ))}
                </List>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Recent Activity */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardHeader title="Recent Activity" />
            <CardContent>
              <List>
                {recentActivity.map((activity, index) => (
                  <React.Fragment key={index}>
                    <ListItem>
                      <ListItemText
                        primary={activity.message}
                        secondary={new Date(activity.timestamp).toLocaleString()}
                      />
                      <Chip
                        label={activity.type.toUpperCase()}
                        color={
                          activity.type === 'success' ? 'success' :
                            activity.type === 'error' ? 'error' :
                              activity.type === 'warning' ? 'warning' : 'info'
                        }
                        size="small"
                      />
                    </ListItem>
                    {index < recentActivity.length - 1 && <Divider />}
                  </React.Fragment>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard; 