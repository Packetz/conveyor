import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  IconButton,
  Tooltip,
  Grid,
  Card,
  CardContent,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Button,
  CircularProgress,
  Divider
} from '@mui/material';
import {
  FilterList,
  Search,
  Download,
  Visibility,
  BarChart as ChartIcon,
  Shield,
  BugReport
} from '@mui/icons-material';
import axios from 'axios';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer } from 'recharts';

// Types for security scan data
interface Finding {
  ruleId: string;
  severity: string;
  description: string;
  location: string;
  lineNumber: number;
  context: string;
  remediation?: string;
}

interface ScanSummary {
  totalFiles: number;
  filesScanned: number;
  filesSkipped: number;
  totalFindings: number;
  findingsBySeverity: Record<string, number>;
  passedCheck: boolean;
}

interface Component {
  name: string;
  version: string;
  type: string;
  license: string;
  source: string;
  vulnerabilities?: Vulnerability[];
}

interface Vulnerability {
  id: string;
  severity: string;
  cvss: number;
  description: string;
  fixedIn?: string;
}

interface SBOM {
  components: Component[];
  format: string;
  version: string;
}

interface ScanResult {
  findings: Finding[];
  summary: ScanSummary;
  sbom?: SBOM;
  timestamp: string;
  environment: string;
  duration: string;
}

// Define color scheme for severity levels
const severityColors: Record<string, string> = {
  'CRITICAL': '#7B1FA2', // Purple
  'HIGH': '#D32F2F',     // Red
  'MEDIUM': '#FF9800',   // Orange
  'LOW': '#4CAF50',      // Green
  'INFO': '#2196F3'      // Blue
};

// Security Report Component
const SecurityReport: React.FC<{ pipelineId?: string }> = ({ pipelineId }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [scanResults, setScanResults] = useState<ScanResult | null>(null);
  const [filteredFindings, setFilteredFindings] = useState<Finding[]>([]);
  const [page, setPage] = useState<number>(0);
  const [rowsPerPage, setRowsPerPage] = useState<number>(10);
  const [severityFilter, setSeverityFilter] = useState<string>('ALL');
  const [typeFilter, setTypeFilter] = useState<string>('ALL');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [showChart, setShowChart] = useState<boolean>(true);

  // Fetch scan results
  useEffect(() => {
    const fetchScanResults = async () => {
      try {
        setLoading(true);
        // In a real app, this would be a call to the API with the pipelineId
        const response = await axios.get(`/api/pipelines/${pipelineId}/security-scan`);
        setScanResults(response.data);
        setFilteredFindings(response.data.findings);
      } catch (error) {
        console.error('Error fetching security scan results:', error);
        // For demo purposes, let's create mock data when the API call fails
        const mockData = generateMockData();
        setScanResults(mockData);
        setFilteredFindings(mockData.findings);
      } finally {
        setLoading(false);
      }
    };

    fetchScanResults();
  }, [pipelineId]);

  // Apply filters
  useEffect(() => {
    if (!scanResults) return;

    let results = [...scanResults.findings];

    // Apply severity filter
    if (severityFilter !== 'ALL') {
      results = results.filter(finding => finding.severity === severityFilter);
    }

    // Apply type filter (assuming ruleId prefix indicates type)
    if (typeFilter !== 'ALL') {
      results = results.filter(finding => finding.ruleId.startsWith(typeFilter));
    }

    // Apply search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      results = results.filter(finding =>
        finding.description.toLowerCase().includes(query) ||
        finding.location.toLowerCase().includes(query) ||
        finding.ruleId.toLowerCase().includes(query)
      );
    }

    setFilteredFindings(results);
    setPage(0); // Reset to first page when filters change
  }, [scanResults, severityFilter, typeFilter, searchQuery]);

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleDownloadReport = () => {
    if (!scanResults) return;

    const element = document.createElement('a');
    const file = new Blob([JSON.stringify(scanResults, null, 2)], { type: 'application/json' });
    element.href = URL.createObjectURL(file);
    element.download = `security-report-${new Date().toISOString()}.json`;
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
  };

  const resetFilters = () => {
    setSeverityFilter('ALL');
    setTypeFilter('ALL');
    setSearchQuery('');
  };

  // Prepare data for charts
  const getSeverityChartData = () => {
    if (!scanResults) return [];

    return Object.entries(scanResults.summary.findingsBySeverity).map(([severity, count]) => ({
      name: severity,
      value: count
    }));
  };

  const getFindingTypeChartData = () => {
    if (!scanResults) return [];

    const typeCounts: Record<string, number> = {};

    scanResults.findings.forEach(finding => {
      const type = finding.ruleId.split('-')[0];
      typeCounts[type] = (typeCounts[type] || 0) + 1;
    });

    return Object.entries(typeCounts).map(([type, count]) => ({
      name: type,
      value: count
    }));
  };

  // Fallback for when API is not available - generate mock data
  const generateMockData = (): ScanResult => {
    return {
      findings: [
        {
          ruleId: "SECRET-001",
          severity: "CRITICAL",
          description: "AWS Access Key ID detected",
          location: "config/settings.js",
          lineNumber: 42,
          context: "const awsKey = 'AKIA[REDACTED]';",
          remediation: "Remove AWS Access Key from code and use environment variables or AWS IAM roles"
        },
        {
          ruleId: "CODE-002",
          severity: "HIGH",
          description: "Potential SQL injection vulnerability",
          location: "src/controllers/users.js",
          lineNumber: 87,
          context: "const query = `SELECT * FROM users WHERE id = ${userId}`;",
          remediation: "Use parameterized queries or prepared statements for database operations"
        },
        {
          ruleId: "VULN-001",
          severity: "HIGH",
          description: "Axios before 0.21.2 allows server-side request forgery (CVE-2021-3749)",
          location: "package.json",
          lineNumber: 15,
          context: "\"axios\": \"^0.21.1\"",
          remediation: "Update axios to version 0.21.2 or later"
        },
        {
          ruleId: "CODE-003",
          severity: "MEDIUM",
          description: "Hardcoded IP address",
          location: "src/services/api.js",
          lineNumber: 12,
          context: "const API_HOST = '192.168.1.100';",
          remediation: "Move IP addresses to configuration files or environment variables"
        },
        {
          ruleId: "CODE-001",
          severity: "HIGH",
          description: "Use of insecure random number generator",
          location: "src/utils/crypto.js",
          lineNumber: 8,
          context: "const token = Math.random().toString(36).substring(2);",
          remediation: "Use a cryptographically secure random number generator"
        },
        {
          ruleId: "LICENSE-001",
          severity: "MEDIUM",
          description: "Detected GPL-3.0 license which may conflict with project requirements",
          location: "node_modules/some-gpl-lib/LICENSE",
          lineNumber: 1,
          context: "GNU General Public License v3.0",
          remediation: "Review license compatibility with legal team"
        }
      ],
      summary: {
        totalFiles: 120,
        filesScanned: 98,
        filesSkipped: 22,
        totalFindings: 6,
        findingsBySeverity: {
          "CRITICAL": 1,
          "HIGH": 3,
          "MEDIUM": 2,
          "LOW": 0,
          "INFO": 0
        },
        passedCheck: false
      },
      sbom: {
        components: [
          {
            name: "axios",
            version: "0.21.1",
            type: "npm",
            license: "MIT",
            source: "https://www.npmjs.com/package/axios",
            vulnerabilities: [
              {
                id: "CVE-2021-3749",
                severity: "HIGH",
                cvss: 8.1,
                description: "Axios before 0.21.2 allows server-side request forgery",
                fixedIn: "0.21.2"
              }
            ]
          },
          {
            name: "lodash",
            version: "4.17.20",
            type: "npm",
            license: "MIT",
            source: "https://www.npmjs.com/package/lodash"
          }
        ],
        format: "cyclonedx",
        version: "1.0"
      },
      timestamp: new Date().toISOString(),
      environment: "development",
      duration: "5.2s"
    };
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="50vh">
        <CircularProgress />
        <Typography variant="h6" sx={{ ml: 2 }}>
          Loading security scan results...
        </Typography>
      </Box>
    );
  }

  if (!scanResults) {
    return (
      <Box textAlign="center" p={4}>
        <Shield sx={{ fontSize: 60, color: 'text.secondary' }} />
        <Typography variant="h5" color="text.secondary">
          No security scan data available
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          Security Scan Results
        </Typography>
        <Box>
          <Tooltip title="Toggle Charts">
            <IconButton onClick={() => setShowChart(!showChart)}>
              <ChartIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Download Report">
            <IconButton onClick={handleDownloadReport}>
              <Download />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Summary Cards */}
      <Grid container spacing={3} mb={4}>
        <Grid item xs={12} md={3}>
          <Card
            sx={{
              bgcolor: scanResults.summary.passedCheck ? 'success.light' : 'error.light',
              color: 'white'
            }}
          >
            <CardContent>
              <Typography variant="h5" fontWeight="bold">
                {scanResults.summary.passedCheck ? 'PASSED' : 'FAILED'}
              </Typography>
              <Typography variant="body2">
                Security Check Status
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h5" fontWeight="bold">
                {scanResults.summary.totalFindings}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Total Findings
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h5" fontWeight="bold">
                {scanResults.summary.filesScanned}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Files Scanned
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h5" fontWeight="bold">
                {scanResults.duration}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Scan Duration
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Charts */}
      {showChart && (
        <Grid container spacing={3} mb={4}>
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Findings by Severity
              </Typography>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={getSeverityChartData()}
                    cx="50%"
                    cy="50%"
                    outerRadius={100}
                    fill="#8884d8"
                    dataKey="value"
                    nameKey="name"
                    label={({ name, value }) => `${name}: ${value}`}
                  >
                    {getSeverityChartData().map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={severityColors[entry.name] || '#8884d8'}
                      />
                    ))}
                  </Pie>
                  <RechartsTooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </Paper>
          </Grid>
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Findings by Type
              </Typography>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={getFindingTypeChartData()}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <RechartsTooltip />
                  <Legend />
                  <Bar dataKey="value" fill="#2196F3" name="Count" />
                </BarChart>
              </ResponsiveContainer>
            </Paper>
          </Grid>
        </Grid>
      )}

      {/* Filters */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          <FilterList sx={{ verticalAlign: 'middle', mr: 1 }} />
          Filters
        </Typography>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Severity</InputLabel>
              <Select
                value={severityFilter}
                label="Severity"
                onChange={(e) => setSeverityFilter(e.target.value)}
              >
                <MenuItem value="ALL">All Severities</MenuItem>
                <MenuItem value="CRITICAL">Critical</MenuItem>
                <MenuItem value="HIGH">High</MenuItem>
                <MenuItem value="MEDIUM">Medium</MenuItem>
                <MenuItem value="LOW">Low</MenuItem>
                <MenuItem value="INFO">Info</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Type</InputLabel>
              <Select
                value={typeFilter}
                label="Type"
                onChange={(e) => setTypeFilter(e.target.value)}
              >
                <MenuItem value="ALL">All Types</MenuItem>
                <MenuItem value="SECRET">Secrets</MenuItem>
                <MenuItem value="CODE">Code Issues</MenuItem>
                <MenuItem value="VULN">Vulnerabilities</MenuItem>
                <MenuItem value="LICENSE">License Issues</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={4}>
            <TextField
              fullWidth
              size="small"
              label="Search"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              InputProps={{
                startAdornment: <Search sx={{ color: 'action.active', mr: 1 }} />,
              }}
            />
          </Grid>
          <Grid item xs={12} md={2}>
            <Button
              fullWidth
              variant="outlined"
              onClick={resetFilters}
            >
              Reset
            </Button>
          </Grid>
        </Grid>
      </Paper>

      {/* Findings Table */}
      <Paper>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Severity</TableCell>
                <TableCell>Rule ID</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>Location</TableCell>
                <TableCell>Line</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredFindings
                .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                .map((finding, index) => (
                  <TableRow key={index} hover>
                    <TableCell>
                      <Chip
                        label={finding.severity}
                        size="small"
                        sx={{
                          bgcolor: severityColors[finding.severity],
                          color: 'white',
                          fontWeight: 'bold'
                        }}
                      />
                    </TableCell>
                    <TableCell>{finding.ruleId}</TableCell>
                    <TableCell>{finding.description}</TableCell>
                    <TableCell>
                      <Tooltip title={finding.location}>
                        <Typography variant="body2" noWrap sx={{ maxWidth: 200 }}>
                          {finding.location}
                        </Typography>
                      </Tooltip>
                    </TableCell>
                    <TableCell>{finding.lineNumber}</TableCell>
                    <TableCell>
                      <Tooltip title="View Details">
                        <IconButton size="small">
                          <Visibility fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="View in Code">
                        <IconButton size="small">
                          <BugReport fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))}
              {filteredFindings.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 3 }}>
                    <Typography variant="body1" color="text.secondary">
                      No findings match your filters
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={filteredFindings.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>

      {/* SBOM Section */}
      {scanResults.sbom && (
        <Box mt={4}>
          <Typography variant="h5" gutterBottom>
            Software Bill of Materials (SBOM)
          </Typography>
          <Paper>
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Component</TableCell>
                    <TableCell>Version</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell>License</TableCell>
                    <TableCell>Vulnerabilities</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {scanResults.sbom.components.map((component, index) => (
                    <TableRow key={index}>
                      <TableCell>{component.name}</TableCell>
                      <TableCell>{component.version}</TableCell>
                      <TableCell>{component.type}</TableCell>
                      <TableCell>{component.license}</TableCell>
                      <TableCell>
                        {component.vulnerabilities && component.vulnerabilities.length > 0 ? (
                          <Chip
                            label={`${component.vulnerabilities.length} found`}
                            color="error"
                            size="small"
                          />
                        ) : (
                          <Chip
                            label="None"
                            color="success"
                            size="small"
                          />
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Box>
      )}

      <Box mt={3} display="flex" justifyContent="flex-end">
        <Typography variant="body2" color="text.secondary">
          Scan completed on {new Date(scanResults.timestamp).toLocaleString()}
        </Typography>
      </Box>
    </Box>
  );
};

export default SecurityReport; 