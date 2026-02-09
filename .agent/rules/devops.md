# DevOps Engineering Rules

## Core Philosophy

- Automate everything  
- Infrastructure as Code first  
- Reproducibility over convenience  
- Zero manual production changes  
- Fail fast, recover faster  

---

## CI/CD Standards

- Every service must have:
  - Build pipeline  
  - Test pipeline  
  - Security scan  
  - Deployment pipeline  

- Pipelines must be:
  - Idempotent  
  - Versioned  
  - Fully automated  

---

## Infrastructure Rules

- All infrastructure defined as code  
- No manual cloud console changes  
- Changes only through pull requests  
- Environments must be consistent  

---

## Security

- Secrets never stored in repositories  
- IAM least privilege  
- Regular vulnerability scanning  
- Audit logging enabled  
- Network segmentation required  

---

## Monitoring

- Logs centralized  
- Alerts for critical systems  
- SLOs and SLAs defined  
- Metrics collected by default  

---

## Reliability

- Automated backups  
- Disaster recovery plan  
- Rollback procedures  
- Blue/green or canary deployments  

---

## Configuration

- Configs separated from code  
- Environment-based configuration  
- Feature flags preferred  
- No hardcoded values  

---

## Documentation

- Runbooks required  
- Incident response documented  
- Clear onboarding guides  
- Architecture diagrams maintained  
