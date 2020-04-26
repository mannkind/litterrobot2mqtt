namespace LitterRobot.Models.Shared
{
    /// <summary>
    /// The shared resource across the application
    /// </summary>
    public class Resource
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LitterRobotId { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LitterRobotSerial { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string PowerStatus { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string UnitStatus { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string UnitStatusText { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LitterRobotNickname { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool Power { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool Cycle { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool PanelLockActive { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool NightLightActive { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool DidNotifyOffline { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool DFITriggered { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool SleepModeActive { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string SleepMode { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CycleCount { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CycleCapacity { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CyclesAfterDrawerFull { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long DFICycleCount { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CleanCycleWaitTimeMinutes { get; set; } = long.MinValue;

        /// <summary>
        ///  
        /// </summary>
        public const string ON_ONE = "1";

        /// <summary>
        /// 
        /// </summary>
        public const string CC = "CC";
    }
}
