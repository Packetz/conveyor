import React, { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Breadcrumbs,
  Link,
  Tab,
  Tabs,
  Paper,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  FormControlLabel,
  Checkbox
} from '@mui/material';
import { NavigateNext, PlayArrow, Refresh, Settings } from '@mui/icons-material';
import { useParams, Link as RouterLink } from 'react-router-dom';
import SecurityReport from '../components/SecurityReport';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`security-tabpanel-${index}`}
      aria-labelledby={`security-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

function a11yProps(index: number) {
  return {
    id: `security-tab-${index}`,
    'aria-controls': `security-tabpanel-${index}`,
  };
}

const SecurityScan: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [tabValue, setTabValue] = useState(0);
  const [openDialog, setOpenDialog] = useState(false);
  const [scanConfig, setScanConfig] = useState({
    scanTypes: ['secret', 'vulnerability', 'code', 'license'],
    severityThreshold: 'HIGH',
    failOnViolation: true,
    generateSBOM: true
  });

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleDialogOpen = () => {
    setOpenDialog(true);
  };

  const handleDialogClose = () => {
    setOpenDialog(false);
  };

  const handleScanTypeChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setScanConfig({
      ...scanConfig,
      scanTypes: event.target.value as string[]
    });
  };

  const handleThresholdChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setScanConfig({
      ...scanConfig,
      severityThreshold: event.target.value as string
    });
  };

  const handleCheckboxChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setScanConfig({
      ...scanConfig,
      [event.target.name]: event.target.checked
    });
  };

  const runSecurityScan = () => {
    // This would make an API call to trigger a security scan
    console.log(`Running security scan with config:`, scanConfig);
    handleDialogClose();
    // In a real app, we would then poll for results or use WebSockets to get updates
  };

  return (
    <Box sx={{ p: 3 }}>
      {/* Breadcrumbs */}
      <Breadcrumbs
        separator={<NavigateNext fontSize="small" />}
        aria-label="breadcrumb"
        sx={{ mb: 3 }}
      >
        <Link component={RouterLink} to="/pipelines" color="inherit">
          Pipelines
        </Link>
        {id && (
          <Link component={RouterLink} to={`/pipelines/${id}`} color="inherit">
            Pipeline {id}
          </Link>
        )}
        <Typography color="text.primary">Security</Typography>
      </Breadcrumbs>

      {/* Header */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          Security Scanning
        </Typography>
        <Box>
          <Button
            variant="contained"
            color="primary"
            startIcon={<PlayArrow />}
            onClick={handleDialogOpen}
            sx={{ mr: 1 }}
          >
            Run Scan
          </Button>
          <Button
            variant="outlined"
            startIcon={<Settings />}
          >
            Configure
          </Button>
        </Box>
      </Box>

      {/* Tabs */}
      <Paper sx={{ mb: 4 }}>
        <Tabs
          value={tabValue}
          onChange={handleTabChange}
          indicatorColor="primary"
          textColor="primary"
        >
          <Tab label="Latest Scan" {...a11yProps(0)} />
          <Tab label="History" {...a11yProps(1)} />
          <Tab label="Configuration" {...a11yProps(2)} />
          <Tab label="Integrations" {...a11yProps(3)} />
        </Tabs>
      </Paper>

      {/* Tab Content */}
      <TabPanel value={tabValue} index={0}>
        <SecurityReport pipelineId={id} />
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        <Typography variant="h6" gutterBottom>
          Scan History
        </Typography>
        <Typography>
          This tab would display historical security scan results and trends over time.
        </Typography>
      </TabPanel>

      <TabPanel value={tabValue} index={2}>
        <Typography variant="h6" gutterBottom>
          Security Scan Configuration
        </Typography>
        <Typography>
          This tab would allow detailed configuration of security scanning rules, thresholds, and schedules.
        </Typography>
      </TabPanel>

      <TabPanel value={tabValue} index={3}>
        <Typography variant="h6" gutterBottom>
          Security Integrations
        </Typography>
        <Typography>
          This tab would allow configuration of integrations with other security tools and notification systems.
        </Typography>
      </TabPanel>

      {/* Run Scan Dialog */}
      <Dialog open={openDialog} onClose={handleDialogClose} maxWidth="sm" fullWidth>
        <DialogTitle>Run Security Scan</DialogTitle>
        <DialogContent>
          <Box sx={{ my: 2 }}>
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel id="scan-types-label">Scan Types</InputLabel>
              <Select
                labelId="scan-types-label"
                multiple
                value={scanConfig.scanTypes}
                onChange={handleScanTypeChange}
                label="Scan Types"
                renderValue={(selected) => (selected as string[]).join(', ')}
              >
                <MenuItem value="secret">Secret Detection</MenuItem>
                <MenuItem value="vulnerability">Vulnerability Scanning</MenuItem>
                <MenuItem value="code">Code Analysis</MenuItem>
                <MenuItem value="license">License Compliance</MenuItem>
              </Select>
            </FormControl>

            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel id="severity-threshold-label">Severity Threshold</InputLabel>
              <Select
                labelId="severity-threshold-label"
                value={scanConfig.severityThreshold}
                onChange={handleThresholdChange}
                label="Severity Threshold"
              >
                <MenuItem value="CRITICAL">Critical</MenuItem>
                <MenuItem value="HIGH">High</MenuItem>
                <MenuItem value="MEDIUM">Medium</MenuItem>
                <MenuItem value="LOW">Low</MenuItem>
                <MenuItem value="INFO">Info</MenuItem>
              </Select>
            </FormControl>

            <FormControlLabel
              control={
                <Checkbox
                  checked={scanConfig.failOnViolation}
                  onChange={handleCheckboxChange}
                  name="failOnViolation"
                />
              }
              label="Fail pipeline on violations above threshold"
            />

            <FormControlLabel
              control={
                <Checkbox
                  checked={scanConfig.generateSBOM}
                  onChange={handleCheckboxChange}
                  name="generateSBOM"
                />
              }
              label="Generate Software Bill of Materials (SBOM)"
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDialogClose}>Cancel</Button>
          <Button onClick={runSecurityScan} variant="contained" color="primary">
            Run Scan
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default SecurityScan; 